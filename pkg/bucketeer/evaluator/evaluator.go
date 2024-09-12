package evaluator

//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../../test/mock/$GOPACKAGE/$GOFILE
import (
	"errors"

	evaluation "github.com/bucketeer-io/bucketeer/evaluation"
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

var (
	errEvaluationNotFound = errors.New("evaluation not found")
)

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
	if len(feature.Prerequisites) == 0 {
		return []*ftproto.Feature{feature}, nil
	}
	preFeatures, err := e.getPrerequisiteFeatures(feature)
	if err != nil {
		return nil, err
	}
	return append([]*ftproto.Feature{feature}, preFeatures...), nil
}

// Gets the features specified as prerequisite
func (e *evaluator) getPrerequisiteFeatures(feature *ftproto.Feature) ([]*ftproto.Feature, error) {
	prerequisites := make(map[string]*ftproto.Feature)
	queue := append([]*ftproto.Feature{}, feature)
	for len(queue) > 0 {
		f := queue[0]
		for _, p := range f.Prerequisites {
			preFeature, err := e.featuresCache.Get(p.FeatureId)
			if err != nil {
				return nil, err
			}
			prerequisites[preFeature.Id] = preFeature
			queue = append(queue, preFeature)
		}
		queue = queue[1:]
	}
	ftList := make([]*ftproto.Feature, 0, len(prerequisites))
	for _, v := range prerequisites {
		ftList = append(ftList, v)
	}
	return ftList, nil
}

func (e *evaluator) findEvaluation(evals []*ftproto.Evaluation, id string) (*ftproto.Evaluation, error) {
	for _, e := range evals {
		if e.FeatureId == id {
			return e, nil
		}
	}
	return nil, errEvaluationNotFound
}
