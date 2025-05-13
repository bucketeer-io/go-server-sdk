package bucketeer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	mockcache "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
	mockevent "github.com/bucketeer-io/go-server-sdk/test/mock/event"
	mockprocessor "github.com/bucketeer-io/go-server-sdk/test/mock/processor"
)

var (
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
						Values:   []string{segment.SegmentId},
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
		Id:            "feature-id-4",
		Enabled:       true,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_BOOLEAN,
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

	// Feature type boolean
	ftBoolean = &ftproto.Feature{
		Id:            "feature-id-boolean",
		Enabled:       true,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_BOOLEAN,
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

	// Feature type int
	ftInt = &ftproto.Feature{
		Id:            "feature-id-int",
		Enabled:       true,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_NUMBER,
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-int10-id",
				Value: "10",
			},
			{
				Id:    "variation-int20-id",
				Value: "20",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-int10-id",
			},
		},
		OffVariation: "variation-int20-id",
	}

	// Feature type float
	ftFloat = &ftproto.Feature{
		Id:            "feature-id-float",
		Enabled:       true,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_NUMBER,
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-float10-id",
				Value: "10.11",
			},
			{
				Id:    "variation-float20-id",
				Value: "20.11",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-float10-id",
			},
		},
		OffVariation: "variation-float20-id",
	}

	// Feature type string
	ftString = &ftproto.Feature{
		Id:            "feature-id-string",
		Enabled:       true,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_STRING,
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-string10-id",
				Value: "value 10",
			},
			{
				Id:    "variation-string20-id",
				Value: "value 20",
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-string10-id",
			},
		},
		OffVariation: "variation-string20-id",
	}

	// Feature type JSON
	ftJSON = &ftproto.Feature{
		Id:            "feature-id-json",
		Enabled:       true,
		Tags:          []string{"server"},
		VariationType: ftproto.Feature_JSON,
		Variations: []*ftproto.Variation{
			{
				Id:    "variation-json1-id",
				Value: string(`{"Str": "str1", "Int": "int1"}`),
			},
			{
				Id:    "variation-json2-id",
				Value: string(`{"Str": "str2", "Int": "int2"}`),
			},
		},
		DefaultStrategy: &ftproto.Strategy{
			FixedStrategy: &ftproto.FixedStrategy{
				Variation: "variation-json1-id",
			},
		},
		OffVariation: string(`{"Str": "str2", "Int": "int2"}`),
	}

	// Segment
	segment = &ftproto.SegmentUsers{
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
)

func TestLocalBoolVariation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  bool
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftBoolean.Id,
			expected:  false,
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftBoolean,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				// &{feature-id-3:0:user-id-1 feature-id-3 0 user-id-1 variation-true-id 0x140000fcac0 true} (*model.Evaluation)
				eval := &model.Evaluation{
					ID:             "feature-id-boolean:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftBoolean.Id,
					FeatureVersion: ftBoolean.Version,
					VariationID:    ftBoolean.Variations[0].Id,
					VariationValue: ftBoolean.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftBoolean.Id,
			expected:  true,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.BoolVariation(ctx, p.user, p.featureID, false)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalBoolVariationDetail(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  model.BKTEvaluationDetails[bool]
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftBoolean.Id,
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      ftBoolean.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonErrorException,
				VariationValue: false,
				VariationID:    "",
			},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftBoolean,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-boolean:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftBoolean.Id,
					FeatureVersion: ftBoolean.Version,
					VariationID:    ftBoolean.Variations[0].Id,
					VariationValue: ftBoolean.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftBoolean.Id,
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      ftBoolean.Id,
				FeatureVersion: ftBoolean.Version,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: true,
				VariationID:    ftBoolean.Variations[0].Id,
			},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.BoolVariationDetails(ctx, p.user, p.featureID, false)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalIntVariation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  int
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected:  0,
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftInt,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-int:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftInt.Id,
					FeatureVersion: ftInt.Version,
					VariationID:    ftInt.Variations[0].Id,
					VariationValue: ftInt.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected:  10,
		},
		{
			desc: "success: value is float",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftFloat,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-float:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftFloat.Id,
					FeatureVersion: ftFloat.Version,
					VariationID:    ftFloat.Variations[0].Id,
					VariationValue: ftFloat.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected:  10,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.IntVariation(ctx, p.user, p.featureID, 0)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalIntVariationDetail(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  model.BKTEvaluationDetails[int]
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      ftInt.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonErrorException,
				VariationValue: 0,
				VariationID:    "",
			},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftInt,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-int:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftInt.Id,
					FeatureVersion: ftInt.Version,
					VariationID:    ftInt.Variations[0].Id,
					VariationValue: ftInt.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      ftInt.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: 10,
				VariationID:    ftInt.Variations[0].Id,
			},
		},
		{
			desc: "success: value is float",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftFloat,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-float:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftFloat.Id,
					FeatureVersion: ftFloat.Version,
					VariationID:    ftFloat.Variations[0].Id,
					VariationValue: ftFloat.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      ftFloat.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: 10,
				VariationID:    ftFloat.Variations[0].Id,
			},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.IntVariationDetails(ctx, p.user, p.featureID, 0)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalInt64Variation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  int64
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected:  0,
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftInt,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-int:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftInt.Id,
					FeatureVersion: ftInt.Version,
					VariationID:    ftInt.Variations[0].Id,
					VariationValue: ftInt.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected:  10,
		},
		{
			desc: "success: value is float",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftFloat,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-float:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftFloat.Id,
					FeatureVersion: ftFloat.Version,
					VariationID:    ftFloat.Variations[0].Id,
					VariationValue: ftFloat.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected:  10,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.Int64Variation(ctx, p.user, p.featureID, 0)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalInt64VariationDetail(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  model.BKTEvaluationDetails[int64]
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      ftInt.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonErrorException,
				VariationValue: 0,
				VariationID:    "",
			},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftInt,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-int:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftInt.Id,
					FeatureVersion: ftInt.Version,
					VariationID:    ftInt.Variations[0].Id,
					VariationValue: ftInt.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftInt.Id,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      ftInt.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: 10,
				VariationID:    ftInt.Variations[0].Id,
			},
		},
		{
			desc: "success: value is float",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftFloat,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-float:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftFloat.Id,
					FeatureVersion: ftFloat.Version,
					VariationID:    ftFloat.Variations[0].Id,
					VariationValue: ftFloat.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      ftFloat.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: 10,
				VariationID:    ftFloat.Variations[0].Id,
			},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.Int64VariationDetails(ctx, p.user, p.featureID, 0)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalFloat64Variation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  float64
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				internalErr := errors.New("internal error")
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected:  0.0,
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftFloat,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-float:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftFloat.Id,
					FeatureVersion: ftFloat.Version,
					VariationID:    ftFloat.Variations[0].Id,
					VariationValue: ftFloat.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected:  10.11,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.Float64Variation(ctx, p.user, p.featureID, 0)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalFloat64VariationDetail(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  model.BKTEvaluationDetails[float64]
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected: model.BKTEvaluationDetails[float64]{
				FeatureID:      ftFloat.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonErrorException,
				VariationValue: 0,
				VariationID:    "",
			},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftFloat,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-float:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftFloat.Id,
					FeatureVersion: ftFloat.Version,
					VariationID:    ftFloat.Variations[0].Id,
					VariationValue: ftFloat.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftFloat.Id,
			expected: model.BKTEvaluationDetails[float64]{
				FeatureID:      ftFloat.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: 10.11,
				VariationID:    ftFloat.Variations[0].Id,
			},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.Float64VariationDetails(ctx, p.user, p.featureID, 0)
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalStringVariation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  string
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftString.Id,
			expected:  "value default",
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftString,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-string:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftString.Id,
					FeatureVersion: ftString.Version,
					VariationID:    ftString.Variations[0].Id,
					VariationValue: ftString.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftString.Id,
			expected:  "value 10",
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.StringVariation(ctx, p.user, p.featureID, "value default")
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalStringVariationDetail(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  model.BKTEvaluationDetails[string]
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftString.Id,
			expected: model.BKTEvaluationDetails[string]{
				FeatureID:      ftString.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonErrorException,
				VariationValue: "value default",
				VariationID:    "",
			},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftString,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-string:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftString.Id,
					FeatureVersion: ftString.Version,
					VariationID:    ftString.Variations[0].Id,
					VariationValue: ftString.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftString.Id,
			expected: model.BKTEvaluationDetails[string]{
				FeatureID:      ftString.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: "value 10",
				VariationID:    ftString.Variations[0].Id,
			},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			variation := sdk.StringVariationDetails(ctx, p.user, p.featureID, "value default")
			assert.Equal(t, p.expected, variation)
		})
	}
}

func TestLocalJSONVariation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc      string
		setup     func(*sdk, *user.User, string)
		user      *user.User
		featureID string
		expected  interface{}
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftString.Id,
			expected:  &DstStruct{},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftJSON,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-json:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftJSON.Id,
					FeatureVersion: ftJSON.Version,
					VariationID:    ftJSON.Variations[0].Id,
					VariationValue: ftJSON.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ftJSON.Id,
			expected:  &DstStruct{Str: "str1", Int: "int1"},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			dst := &DstStruct{}
			sdk.JSONVariation(ctx, p.user, p.featureID, dst)
			assert.Equal(t, p.expected, dst)
		})
	}
}

func TestLocalObjectVariation(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc         string
		setup        func(*sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue interface{}
		expected     interface{}
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:         &user.User{ID: "user-id-1"},
			featureID:    ftString.Id,
			defaultValue: &DstStruct{Str: "str0", Int: "int0"},
			expected:     &DstStruct{Str: "str0", Int: "int0"},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftJSON,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-json:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftJSON.Id,
					FeatureVersion: ftJSON.Version,
					VariationID:    ftJSON.Variations[0].Id,
					VariationValue: ftJSON.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:         &user.User{ID: "user-id-1"},
			featureID:    ftJSON.Id,
			defaultValue: &DstStruct{Str: "str0", Int: "int0"},
			expected:     &DstStruct{Str: "str1", Int: "int1"},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			value := sdk.ObjectVariation(ctx, p.user, p.featureID, p.defaultValue)
			assert.Equal(t, p.expected, value)
		})
	}
}

func TestLocalObjectVariationDetail(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc         string
		setup        func(*sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue interface{}
		expected     interface{}
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(u, featureID)
			},
			user:         &user.User{ID: "user-id-1"},
			featureID:    ftJSON.Id,
			defaultValue: &DstStruct{"defaultStr", "defaultInt"},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      ftJSON.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonErrorException,
				VariationValue: &DstStruct{"defaultStr", "defaultInt"},
				VariationID:    "",
			},
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					ftJSON,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
				eval := &model.Evaluation{
					ID:             "feature-id-json:0:user-id-1",
					UserID:         "user-id-1",
					FeatureID:      ftJSON.Id,
					FeatureVersion: ftJSON.Version,
					VariationID:    ftJSON.Variations[0].Id,
					VariationValue: ftJSON.Variations[0].Value,
					Reason:         &model.Reason{Type: model.ReasonDefault},
				}
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, eval)
			},
			user:         &user.User{ID: "user-id-1"},
			featureID:    ftJSON.Id,
			defaultValue: &DstStruct{},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      ftJSON.Id,
				FeatureVersion: 0,
				UserID:         "user-id-1",
				Reason:         model.EvaluationReasonDefault,
				VariationValue: &DstStruct{Str: "str1", Int: "int1"},
				VariationID:    ftJSON.Variations[0].Id,
			},
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			value := sdk.ObjectVariationDetails(ctx, p.user, p.featureID, p.defaultValue)
			assert.Equal(t, p.expected, value)
		})
	}
}

func TestGetEvaluationLocally(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")
	patterns := []struct {
		desc        string
		setup       func(*sdk, *user.User, string)
		user        *user.User
		featureID   string
		expected    *model.Evaluation
		expectedErr error
	}{
		{
			desc: "err: internal error",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(featureID).Return(
					nil,
					internalErr,
				)
				e := fmt.Errorf("internal error while evaluating user locally: %w", internalErr)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(e, model.SDKGetEvaluation)
			},
			user:        &user.User{ID: "user-id-1"},
			featureID:   "feature-d-1",
			expected:    nil,
			expectedErr: internalErr,
		},
		{
			desc: "success",
			setup: func(s *sdk, u *user.User, featureID string) {
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(ft3.Id).Return(
					ft3,
					nil,
				)
				s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(ft4.Id).Return(
					ft4,
					nil,
				)
				s.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Get(segment.SegmentId).Return(
					segment,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), model.SDKGetEvaluation,
				)
			},
			user:      &user.User{ID: "user-id-1"},
			featureID: ft3.Id,
			expected: &model.Evaluation{
				ID:             "feature-id-3:0:user-id-1",
				FeatureID:      "feature-id-3",
				FeatureVersion: 0,
				UserID:         "user-id-1",
				VariationID:    "variation-true-id",
				Reason:         &model.Reason{Type: model.ReasonDefault},
				VariationValue: "true",
			},
			expectedErr: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			sdk := newSDKLocalEvaluationWithMock(t, controller)
			ctx := context.Background()
			p.setup(sdk, p.user, p.featureID)
			evaluation, err := sdk.getEvaluation(ctx, model.SourceIDGoServer, p.user, p.featureID)
			assert.Equal(t, p.expected, evaluation)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestCloseProcessor(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	patterns := []struct {
		desc  string
		setup func(context.Context, *sdk)
		isErr bool
	}{
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().Close(ctx).Return(nil)
				s.featureFlagCacheProcessor.(*mockprocessor.MockFeatureFlagProcessor).EXPECT().Close()
				s.segmentUserCacheProcessor.(*mockprocessor.MockSegmentUserProcessor).EXPECT().Close()
			},
			isErr: false,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			s := newSDKLocalEvaluationWithMock(t, mockCtrl)
			ctx := context.Background()
			p.setup(ctx, s)
			assert.NoError(t, s.Close(ctx))
		})
	}
}

func newSDKLocalEvaluationWithMock(t *testing.T, mockCtrl *gomock.Controller) *sdk {
	t.Helper()
	return &sdk{
		tag:                       "server",
		apiClient:                 mockapi.NewMockClient(mockCtrl),
		eventProcessor:            mockevent.NewMockProcessor(mockCtrl),
		enableLocalEvaluation:     true,
		featureFlagCacheProcessor: mockprocessor.NewMockFeatureFlagProcessor(mockCtrl),
		segmentUserCacheProcessor: mockprocessor.NewMockSegmentUserProcessor(mockCtrl),
		featureFlagsCache:         mockcache.NewMockFeaturesCache(mockCtrl),
		segmentUsersCache:         mockcache.NewMockSegmentUsersCache(mockCtrl),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
	}
}
