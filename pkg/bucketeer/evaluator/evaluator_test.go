package evaluator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

var (
	internalErr = errors.New("internal error")
	// Feature 1
	ft1 = &ftproto.Feature{
		Id:      "feature-id-1",
		Enabled: true,
		Tags:    []string{"server"},
		Prerequisites: []*ftproto.Prerequisite{
			{
				FeatureId:   "feature-id-2",
				VariationId: "variation-true-id",
			},
		},
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-true-id",
				Value: "true",
			},
			{
				Id:    "variation-false-id",
				Value: "false",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-true-id",
			},
		},
		OffVariation: "variation-false-id",
	}
	// Feature 2
	ft2 = &ftproto.Feature{
		Id:      "feature-id-2",
		Enabled: true,
		Tags:    []string{"server"},
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-true-id",
				Value: "true",
			},
			{
				Id:    "variation-false-id",
				Value: "false",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-true-id",
			},
		},
		OffVariation: "variation-false-id",
	}
	// Feature 3
	ft3 = &ftproto.Feature{
		Id:      "feature-id-3",
		Enabled: true,
		Tags:    []string{"server"},
		Prerequisites: []*ftproto.Prerequisite{
			{
				FeatureId:   "feature-id-4",
				VariationId: "variation-true-id",
			},
		},
		Rules: []*ftproto.Rule{
			{
				Clauses: []*ftproto.Clause{
					{
						Operator: ftproto.Clause_SEGMENT,
						Values:   []string{segment1.SegmentId},
					},
					{
						Operator: ftproto.Clause_SEGMENT,
						Values:   []string{segment2.SegmentId},
					},
				},
			},
		},
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-true-id",
				Value: "true",
			},
			{
				Id:    "variation-false-id",
				Value: "false",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-true-id",
			},
		},
		OffVariation: "variation-false-id",
	}
	// Feature 4
	ft4 = &ftproto.Feature{
		Id:      "feature-id-4",
		Enabled: false,
		Tags:    []string{"server"},
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-true-id",
				Value: "true",
			},
			{
				Id:    "variation-false-id",
				Value: "false",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-true-id",
			},
		},
		OffVariation: "variation-false-id",
	}
	// Feature 5
	ft5 = &ftproto.Feature{
		Id:      "feature-id-5",
		Enabled: true,
		Tags:    []string{"server"},
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-true-id",
				Value: "true",
			},
			{
				Id:    "variation-false-id",
				Value: "false",
			},
		},
		Rules: []*ftproto.Rule{
			{
				Clauses: []*ftproto.Clause{
					{
						Id:       "clause-id",
						Operator: ftproto.Clause_SEGMENT,
						Values:   []string{"segment-id-2"},
					},
				},
				Strategy: &ftproto.Strategy{
					FixedStrategy: &ftproto.FixedStrategy{
						Variation: "variation-true-id",
					},
				},
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-true-id",
			},
		},
		OffVariation: "variation-false-id",
	}

	// Segments
	segment1 = &ftproto.SegmentUsers{
		SegmentId: "segment-id-1",
		Users: []*ftproto.SegmentUser{
			{
				UserId: "user-id-1",
			},
		},
	}
	segment2 = &ftproto.SegmentUsers{
		SegmentId: "segment-id-2",
		Users: []*ftproto.SegmentUser{
			{
				SegmentId: "segment-id-2",
				State:     ftproto.SegmentUser_INCLUDED,
				UserId:    "user-id-2",
			},
			{
				SegmentId: "segment-id-2",
				State:     ftproto.SegmentUser_INCLUDED,
				UserId:    "user-id-3",
			},
		},
	}

	// Users
	user1 = &user.User{ID: "user-id-1"}
	user2 = &user.User{ID: "user-id-2"}
)

func TestEvalute(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	patterns := []struct {
		desc           string
		setup          func(*evaluator)
		user           *user.User
		featureID, tag string
		expected       *model.Evaluation
		expectedErr    error
	}{
		{
			desc: "err: failed to get feature flag from cache",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft1.Id).Return(nil, internalErr)
			},
			user:        nil,
			featureID:   ft1.Id,
			expected:    nil,
			expectedErr: internalErr,
		},
		{
			desc: "err: failed to get prerequisite feature flag from cache",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft1.Id).Return(ft1, nil)
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(nil, internalErr)
			},
			user:        nil,
			featureID:   ft1.Id,
			expected:    nil,
			expectedErr: internalErr,
		},
		{
			desc: "err: failed to get segment from cache",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft5.Id).Return(ft5, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(nil, internalErr)
			},
			user:        nil,
			featureID:   ft5.Id,
			expected:    nil,
			expectedErr: internalErr,
		},
		{
			desc: "success: with no prerequisites",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft5.Id).Return(ft5, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(segment2, nil)
			},
			user:      user1,
			featureID: ft5.Id,
			tag:       "server",
			expected: &model.Evaluation{
				ID:             "feature-id-5:0:user-id-1",
				FeatureID:      "feature-id-5",
				FeatureVersion: 0,
				UserID:         "user-id-1",
				VariationID:    "variation-true-id",
				Reason:         &model.Reason{Type: model.ReasonDefault},
				VariationValue: "true",
			},
			expectedErr: nil,
		},
		{
			desc: "success: with prerequisite feature disabled (It must return the off variation)",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft3.Id).Return(ft3, nil)
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(ft4, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment1.SegmentId).Return(segment1, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(segment2, nil)
			},
			user:      user1,
			featureID: ft3.Id,
			expected: &model.Evaluation{
				ID:             "feature-id-3:0:user-id-1",
				FeatureID:      "feature-id-3",
				FeatureVersion: 0,
				UserID:         "user-id-1",
				VariationID:    "variation-false-id",
				Reason:         &model.Reason{Type: model.ReasonPrerequisite},
				VariationValue: "false",
			},
			expectedErr: nil,
		},
		{
			desc: "success: with prerequisite feature enabled (It must return the default strategy variation)",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft1.Id).Return(ft1, nil)
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft2.Id).Return(ft2, nil)
			},
			user:      user1,
			featureID: ft1.Id,
			tag:       "server",
			expected: &model.Evaluation{
				ID:             "feature-id-1:0:user-id-1",
				FeatureID:      "feature-id-1",
				FeatureVersion: 0,
				UserID:         "user-id-1",
				VariationID:    "variation-true-id",
				Reason:         &model.Reason{Type: model.ReasonDefault},
				VariationValue: "true",
			},
			expectedErr: nil,
		},
		{
			desc: "success: with segment user",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft5.Id).Return(ft5, nil)
				e.segmentUsersCache.(*mock.MockSegmentUsersCache).EXPECT().Get(segment2.SegmentId).Return(segment2, nil)
			},
			user:      user2,
			featureID: ft5.Id,
			tag:       "server",
			expected: &model.Evaluation{
				ID:             "feature-id-5:0:user-id-2",
				FeatureID:      "feature-id-5",
				FeatureVersion: 0,
				UserID:         "user-id-2",
				VariationID:    "variation-true-id",
				Reason:         &model.Reason{Type: model.ReasonRule},
				VariationValue: "true",
			},
			expectedErr: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			evaluator := newEvaluator(t, p.tag, controller)
			p.setup(evaluator)
			features, err := evaluator.Evaluate(p.user, p.featureID)
			assert.Equal(t, p.expected, features)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestGetTargetFeatures(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	patterns := []struct {
		desc        string
		setup       func(*evaluator)
		feature     *ftproto.Feature
		expected    []*ftproto.Feature
		expectedErr error
	}{
		{
			desc: "err: failed to get feature flag from cache",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(nil, internalErr)
			},
			feature:     ft3,
			expected:    nil,
			expectedErr: internalErr,
		},
		{
			desc: "success",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(
					ft4,
					nil,
				)
			},
			feature:     ft3,
			expected:    []*ftproto.Feature{ft3, ft4},
			expectedErr: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			evaluator := newEvaluator(t, "tag", controller)
			p.setup(evaluator)
			features, err := evaluator.getTargetFeatures(p.feature)
			assert.Equal(t, p.expected, features)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestGetPrerequisiteFeatures(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	patterns := []struct {
		desc        string
		setup       func(*evaluator)
		feature     *ftproto.Feature
		expected    []*ftproto.Feature
		expectedErr error
	}{
		{
			desc: "err: failed to get feature flag from cache",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(nil, internalErr)
			},
			feature:     ft3,
			expected:    nil,
			expectedErr: internalErr,
		},
		{
			desc: "success",
			setup: func(e *evaluator) {
				e.featuresCache.(*mock.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(
					ft4,
					nil,
				)
			},
			feature:     ft3,
			expected:    []*ftproto.Feature{ft4},
			expectedErr: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			evaluator := newEvaluator(t, "tag", controller)
			p.setup(evaluator)
			features, err := evaluator.getPrerequisiteFeatures(p.feature)
			assert.Equal(t, p.expected, features)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestFindEvaluation(t *testing.T) {
	t.Parallel()
	eval := &ftproto.Evaluation{
		FeatureId: "feature-id-1",
	}

	patterns := []struct {
		desc        string
		evaluations []*ftproto.Evaluation
		featureID   string
		expected    *ftproto.Evaluation
		expectedErr error
	}{
		{
			desc:        "err: evaluation not found",
			evaluations: []*ftproto.Evaluation{eval},
			featureID:   "feature-id-2",
			expected:    nil,
			expectedErr: errEvaluationNotFound,
		},
		{
			desc:        "success: evaluation found",
			evaluations: []*ftproto.Evaluation{eval},
			featureID:   "feature-id-1",
			expected:    eval,
			expectedErr: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			eval := &evaluator{}
			evaluation, err := eval.findEvaluation(p.evaluations, p.featureID)
			assert.Equal(t, p.expected, evaluation)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func newEvaluator(t *testing.T, tag string, controller *gomock.Controller) *evaluator {
	t.Helper()
	ffcache := mock.NewMockFeaturesCache(controller)
	sucache := mock.NewMockSegmentUsersCache(controller)
	return &evaluator{
		tag:               tag,
		featuresCache:     ffcache,
		segmentUsersCache: sucache,
	}
}
