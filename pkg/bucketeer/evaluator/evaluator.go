package evaluator

//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
import (
	"errors"

	evaluation "github.com/bucketeer-io/bucketeer/v2/evaluation/go"
	ftdomain "github.com/bucketeer-io/bucketeer/v2/pkg/feature/domain"
	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	userproto "github.com/bucketeer-io/bucketeer/v2/proto/user"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

type EvaluateLocally interface {
	Evaluate(user *user.User, featureID string) (*model.Evaluation, error)
}

type evaluator struct {
	tag                string
	featuresCache      cache.FeaturesCache
	segmentUsersCache  cache.SegmentUsersCache
	holdoutGroupsCache cache.HoldoutGroupsCache
	loggers            *log.Loggers
}

func NewEvaluator(
	tag string,
	featuresCache cache.FeaturesCache,
	segmentUsersCache cache.SegmentUsersCache,
	holdoutGroupsCache cache.HoldoutGroupsCache,
	loggers *log.Loggers,
) EvaluateLocally {
	return &evaluator{
		tag:                tag,
		featuresCache:      featuresCache,
		segmentUsersCache:  segmentUsersCache,
		holdoutGroupsCache: holdoutGroupsCache,
		loggers:            loggers,
	}
}

var ErrEvaluationNotFound = errors.New("evaluation not found")

func (e *evaluator) Evaluate(user *user.User, featureID string) (*model.Evaluation, error) {
	// Get the target feature
	feature, err := e.featuresCache.Get(featureID)
	if err != nil {
		return nil, err
	}
	// Include the prerequisite features if is configured
	targetFeatures, err := e.getTargetFeatures(feature)
	if err != nil {
		return nil, err
	}
	// List and get all the segments if is configured in the targeting rules
	evaluator := evaluation.NewEvaluator()
	sIDs := evaluator.ListSegmentIDs(feature)
	segmentUsers := make(map[string][]*ftproto.SegmentUser, len(sIDs))
	segments := make(map[string]*ftproto.Segment, len(sIDs))
	for _, id := range sIDs {
		su, err := e.segmentUsersCache.Get(id)
		if err != nil {
			return nil, err
		}
		segmentUsers[su.SegmentId] = su.Users
		// The cached SegmentUsers message also carries the segment's rules,
		// so the user can be matched by attribute-based rules as well as by the user list.
		segments[su.SegmentId] = &ftproto.Segment{
			Id:    su.SegmentId,
			Rules: su.Rules,
		}
	}
	// Get the holdout groups that are serving in the environment.
	// The evaluation module serves the holdout variation to the held-out users
	// before it looks at the flag itself.
	holdoutGroups, err := e.holdoutGroupsCache.Get()
	if err != nil {
		return nil, err
	}
	// A group with the SEGMENT assignment reads its members from a segment
	// that no flag rule has to reference, so it is not listed above.
	// Only the user list is needed. The segment rules are not consulted for the holdout membership.
	for _, id := range holdoutSegmentIDs(holdoutGroups, targetFeatures) {
		if _, ok := segmentUsers[id]; ok {
			continue
		}
		su, err := e.segmentUsersCache.Get(id)
		if err != nil {
			if !errors.Is(err, cache.ErrNotFound) {
				return nil, err
			}
			// The flags and the segment users are polled separately,
			// so the group can be cached before its segment.
			// The user is evaluated without the group, which means not held out,
			// instead of failing the evaluation, which would return the default value.
			e.loggers.Errorf(
				"bucketeer/evaluator: the segment of the holdout group is not cached yet. "+
					"The user is evaluated as not held out (featureID: %s, segmentID: %s, userID: %s)",
				featureID,
				id,
				user.ID,
			)
			continue
		}
		segmentUsers[id] = su.Users
	}
	// Convert to evaluation's user proto message
	u := &userproto.User{
		Id:   user.ID,
		Data: user.Data,
	}
	// Evaluate the end user
	userEvaluations, err := evaluator.EvaluateFeatures(
		targetFeatures,
		u,
		segmentUsers,
		segments,
		e.tag,
		holdoutGroups...,
	)
	if err != nil {
		return nil, err
	}
	// Find the target evaluation from the results
	eval, err := e.findEvaluation(userEvaluations.Evaluations, featureID)
	if err != nil {
		return nil, err
	}
	// Convert to SDK's model
	evaluation := &model.Evaluation{
		ID:             eval.Id,
		FeatureID:      eval.FeatureId,
		FeatureVersion: eval.FeatureVersion,
		UserID:         eval.UserId,
		VariationID:    eval.VariationId,
		VariationValue: eval.VariationValue,
		VariationName:  eval.VariationName,
		Reason: &model.Reason{
			Type:           model.ReasonType(eval.Reason.Type.String()),
			RuleID:         eval.Reason.RuleId,
			HoldoutGroupID: eval.Reason.HoldoutGroupId,
		},
	}
	return evaluation, nil
}

// holdoutSegmentIDs returns the segments the evaluation module reads
// to resolve the holdout membership of the user for the given features.
//
// The membership is resolved only when a group targets one of the evaluated features.
// The other evaluations don't read the segment, so it is not loaded for them:
// the cache decodes all the users of the segment every time it is read.
func holdoutSegmentIDs(groups []*ftproto.HoldoutGroup, features []*ftproto.Feature) []string {
	var targeting []*ftproto.HoldoutGroup
	for _, group := range groups {
		// Only a group with the SEGMENT assignment has a segment.
		if group.GetAssignmentType() != ftproto.HoldoutGroup_SEGMENT {
			continue
		}
		if holdoutGroupTargets(group, features) {
			targeting = append(targeting, group)
		}
	}
	return evaluation.HoldoutSegmentIDs(targeting)
}

// holdoutGroupTargets reports whether the group lists one of the features.
//
// It must never be narrower than the evaluation module,
// otherwise a held-out user would be evaluated without the segment.
// So it matches on the feature ID alone.
func holdoutGroupTargets(group *ftproto.HoldoutGroup, features []*ftproto.Feature) bool {
	for _, target := range group.GetFeatures() {
		for _, feature := range features {
			if target.GetFeatureId() == feature.Id {
				return true
			}
		}
	}
	return false
}

func (e *evaluator) getTargetFeatures(feature *ftproto.Feature) ([]*ftproto.Feature, error) {
	// Check if the flag depends on other flags.
	// If not, we return only the target flag
	df := &ftdomain.Feature{Feature: feature}
	preFlagIDs := df.FeatureIDsDependsOn()
	if len(preFlagIDs) == 0 {
		return []*ftproto.Feature{feature}, nil
	}
	preFeatures, err := e.getPrerequisiteFeaturesFromCache(preFlagIDs)
	if err != nil {
		return nil, err
	}
	return append([]*ftproto.Feature{feature}, preFeatures...), nil
}

// Gets the features specified as prerequisite
func (e *evaluator) getPrerequisiteFeaturesFromCache(preFlagIDs []string) ([]*ftproto.Feature, error) {
	// Prerequisites contain dependent flags, which could also contain the same flags.
	// So, it uses a map to deduplicate the flags if needed.
	deduplicateFlagIDs := make(map[string]struct{})
	for _, id := range preFlagIDs {
		deduplicateFlagIDs[id] = struct{}{}
	}
	prerequisites := make([]*ftproto.Feature, 0, len(deduplicateFlagIDs))
	for id := range deduplicateFlagIDs {
		preFeature, err := e.featuresCache.Get(id)
		if err != nil {
			return nil, err
		}
		prerequisites = append(prerequisites, preFeature)
	}
	return prerequisites, nil
}

func (e *evaluator) findEvaluation(evals []*ftproto.Evaluation, id string) (*ftproto.Evaluation, error) {
	for _, e := range evals {
		if e.FeatureId == id {
			return e, nil
		}
	}
	return nil, ErrEvaluationNotFound
}
