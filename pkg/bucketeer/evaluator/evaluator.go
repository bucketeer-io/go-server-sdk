package evaluator

//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
import (
	"errors"

	evaluation "github.com/bucketeer-io/bucketeer/evaluation/go"
	ftdomain "github.com/bucketeer-io/bucketeer/pkg/feature/domain"
	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	userproto "github.com/bucketeer-io/bucketeer/proto/user"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

type EvaluateLocally interface {
	Evaluate(user *user.User, featureID string) (*model.Evaluation, error)
}

type evaluator struct {
	tag               string
	featuresCache     cache.FeaturesCache
	segmentUsersCache cache.SegmentUsersCache
}

func NewEvaluator(
	tag string,
	featuresCache cache.FeaturesCache,
	segmentUsersCache cache.SegmentUsersCache,
) EvaluateLocally {
	return &evaluator{
		tag:               tag,
		featuresCache:     featuresCache,
		segmentUsersCache: segmentUsersCache,
	}
}

var errEvaluationNotFound = errors.New("evaluation not found")

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
	segments := make(map[string][]*ftproto.SegmentUser, len(sIDs))
	for _, id := range sIDs {
		segment, err := e.segmentUsersCache.Get(id)
		if err != nil {
			return nil, err
		}
		segments[segment.SegmentId] = segment.Users
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
		segments,
		e.tag,
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
			Type:   model.ReasonType(eval.Reason.Type.String()),
			RuleID: eval.Reason.RuleId,
		},
	}
	return evaluation, nil
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
	return nil, errEvaluationNotFound
}
