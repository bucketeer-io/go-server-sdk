package bucketeer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
	mockevent "github.com/bucketeer-io/go-server-sdk/test/mock/event"
)

const (
	sdkTag       = "go-server"
	sdkVersion   = "1.5.5"
	sdkUserID    = "user-id"
	sdkFeatureID = "feature-id"
	sdkGoalID    = "goal-id"
)

func TestBoolVariation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue bool
		expected     bool
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected:     false,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected:     false,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "true")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.BoolVariation(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBoolVariationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue bool
		expected     model.BKTEvaluationDetails[bool]
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: false,
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorException,
			},
		},
		{
			desc: "return default value when filed to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: false,
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorWrongType,
			},
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "true")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationValue: true,
				VariationName:  "testVersionName",
				Reason:         model.EvaluationReasonTarget,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.BoolVariationDetails(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIntVariation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue int
		expected     int
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "return default value when failed to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 10, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					10,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 20, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					20,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     2,
		},
		{
			desc: "success: value is float",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 20, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					20,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.IntVariation(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIntVariationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue int
		expected     model.BKTEvaluationDetails[int]
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationName:  "",
				VariationValue: 1,
				Reason:         model.EvaluationReasonErrorException,
			},
		},
		{
			desc: "return default value when failed to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 10, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					10,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationName:  "",
				VariationValue: 1,
				Reason:         model.EvaluationReasonErrorWrongType,
			},
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 20, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					20,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: 2,
				Reason:         model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success: value is float",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 20, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					20,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: 2,
				Reason:         model.EvaluationReasonTarget,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.IntVariationDetails(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInt64Variation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue int64
		expected     int64
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100,
					err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 5, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					5,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 10, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					10,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     2,
		},
		{
			desc: "success: value is float",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 10, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					10,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.Int64Variation(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestInt64VariationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue int64
		expected     model.BKTEvaluationDetails[int64]
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100,
					err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: 1,
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorException,
			},
		},
		{
			desc: "return default value when failed to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 5, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					5,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: 1,
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorWrongType,
			},
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 10, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					10,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationValue: 2,
				VariationName:  "testVersionName",
				Reason:         model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success: value is float",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 10, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					10,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected: model.BKTEvaluationDetails[int64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationValue: 2,
				VariationName:  "testVersionName",
				Reason:         model.EvaluationReasonTarget,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.Int64VariationDetails(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFloat64Variation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue float64
		expected     float64
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected:     1.1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected:     1.1,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected:     2.2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.Float64Variation(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFloat64VariationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue float64
		expected     model.BKTEvaluationDetails[float64]
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected: model.BKTEvaluationDetails[float64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: 1.1,
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorException,
			},
		},
		{
			desc: "return default value when failed to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected: model.BKTEvaluationDetails[float64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: 1.1,
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorWrongType,
			},
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected: model.BKTEvaluationDetails[float64]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationValue: 2.2,
				VariationName:  "testVersionName",
				Reason:         model.EvaluationReasonTarget,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.Float64VariationDetails(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringVariation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue string
		expected     string
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected:     "default",
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected:     "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.StringVariation(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestStringVariationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue string
		expected     model.BKTEvaluationDetails[string]
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected: model.BKTEvaluationDetails[string]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: "default",
				VariationName:  "",
				Reason:         model.EvaluationReasonErrorException,
			},
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected: model.BKTEvaluationDetails[string]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: "value",
				Reason:         model.EvaluationReasonTarget,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := s.StringVariationDetails(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestJSONVariation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	tests := []struct {
		desc      string
		setup     func(context.Context, *sdk, *user.User, string)
		user      *user.User
		featureID string
		dst       interface{}
		expected  interface{}
	}{
		{
			desc: "failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			dst:       &DstStruct{},
			expected:  &DstStruct{},
		},
		{
			desc: "faled to unmarshal variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `invalid`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			dst:       &DstStruct{},
			expected:  &DstStruct{},
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "int": "int2"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			dst:       &DstStruct{},
			expected:  &DstStruct{Str: "str2", Int: "int2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			s.JSONVariation(ctx, tt.user, tt.featureID, tt.dst)
			assert.Equal(t, tt.expected, tt.dst)
		})
	}
}

func TestObjectVariation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue interface{}
		expected     interface{}
	}{
		{
			desc: "failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: DstStruct{
				Str: "str1",
				Int: "int1",
			},
			expected: DstStruct{
				Str: "str1",
				Int: "int1",
			},
		},
		{
			desc: "failed to unmarshal variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `invalid`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: DstStruct{
				Str: "str1",
				Int: "int1",
			},
			expected: DstStruct{
				Str: "str1",
				Int: "int1",
			},
		},
		{
			desc: "success:map",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "int": "int2"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: map[string]interface{}{"str": "str0", "int": "int0"},
			expected:     map[string]interface{}{"str": "str2", "int": "int2"},
		},
		{
			desc: "success:bool",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `true`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected:     true,
		},
		{
			desc: "success:array of string",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `["str1", "str2"]`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: []interface{}{"str0", "str0"},
			expected:     []interface{}{"str1", "str2"},
		},
		{
			desc: "success:array of float",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "results": [1.1,2.22]}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: map[string]interface{}{"str": "str0", "results": []interface{}{float64(1.5), float64(2.5)}},
			expected:     map[string]interface{}{"str": "str2", "results": []interface{}{float64(1.1), float64(2.22)}},
		},
		{
			desc: "success:array of object",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `[{"str": "str1", "results": [1,2]}, {"str": "str2", "results": [3,4]}]`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: []interface{}{
				map[string]interface{}{"str": "str0", "results": []interface{}{float64(0), float64(0)}},
				map[string]interface{}{"str": "str0", "results": []interface{}{float64(0), float64(0)}},
			},
			expected: []interface{}{
				map[string]interface{}{"str": "str1", "results": []interface{}{float64(1), float64(2)}},
				map[string]interface{}{"str": "str2", "results": []interface{}{float64(3), float64(4)}},
			},
		},
		{
			desc: "success: json",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str1", "int": "int1"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: &DstStruct{
				Str: "str0",
				Int: "int0",
			},
			expected: &DstStruct{
				Str: "str1",
				Int: "int1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			value := s.ObjectVariation(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestObjectVariationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	type DstStruct struct {
		Str string `json:"str"`
		Int string `json:"int"`
	}
	tests := []struct {
		desc         string
		setup        func(context.Context, *sdk, *user.User, string)
		user         *user.User
		featureID    string
		defaultValue interface{}
		expected     model.BKTEvaluationDetails[interface{}]
	}{
		{
			desc: "failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusInternalServerError)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorException,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: DstStruct{
				Str: "str1",
				Int: "int1",
			},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: DstStruct{
					Str: "str1",
					Int: "int1",
				},
				VariationName: "",
				Reason:        model.EvaluationReasonErrorException,
			},
		},
		{
			desc: "failed to unmarshal variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `invalid`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					user,
					featureID,
					model.ReasonErrorWrongType,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: DstStruct{
				Str: "str1",
				Int: "int1",
			},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 0,
				UserID:         sdkUserID,
				VariationID:    "",
				VariationValue: DstStruct{
					Str: "str1",
					Int: "int1",
				},
				VariationName: "",
				Reason:        model.EvaluationReasonErrorWrongType,
			},
		},
		{
			desc: "success:map",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "int": "int2"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: map[string]interface{}{"str": "str0", "int": "int0"},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: map[string]interface{}{"str": "str2", "int": "int2"},
				Reason:         model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success:bool",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `true`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: true,
				Reason:         model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success:array of string",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `["str1", "str2"]`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: []interface{}{"str0", "str0"},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: []interface{}{"str1", "str2"},
				Reason:         model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success:array of float",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "results": [1,2]}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: map[string]interface{}{"str": "str0", "results": []interface{}{float64(0), float64(0)}},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: map[string]interface{}{
					"str": "str2", "results": []interface{}{float64(1), float64(2)},
				},
				Reason: model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success:array of object",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `[{"str": "str1", "results": [1,2]}, {"str": "str2", "results": [3,4]}]`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: []interface{}{
				map[string]interface{}{"str": "str0", "results": []interface{}{float64(0), float64(0)}},
				map[string]interface{}{"str": "str0", "results": []interface{}{float64(0), float64(0)}},
			},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: []interface{}{
					map[string]interface{}{"str": "str1", "results": []interface{}{float64(1), float64(2)}},
					map[string]interface{}{"str": "str2", "results": []interface{}{float64(3), float64(4)}},
				},
				Reason: model.EvaluationReasonTarget,
			},
		},
		{
			desc: "success: json",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, `{"str": "str1", "int": "int1"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			defaultValue: &DstStruct{
				Str: "str0",
				Int: "int0",
			},
			expected: model.BKTEvaluationDetails[interface{}]{
				FeatureID:      sdkFeatureID,
				FeatureVersion: 1,
				UserID:         sdkUserID,
				VariationID:    "testVersion",
				VariationName:  "testVersionName",
				VariationValue: &DstStruct{
					Str: "str1",
					Int: "int1",
				},
				Reason: model.EvaluationReasonTarget,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			value := s.ObjectVariationDetails(ctx, tt.user, tt.featureID, tt.defaultValue)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestGetEvaluation(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc          string
		setup         func(context.Context, *sdk, *user.User, string)
		user          *user.User
		featureID     string
		expectedValue string
		isErr         bool
	}{
		{
			desc:          "invalid user",
			setup:         nil,
			user:          newUser(t, ""),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "get evaluations returns timeout error",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusGatewayTimeout)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "get evaluations returns internal error",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				err := api.NewErrStatus(http.StatusGatewayTimeout)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil,
					100, err,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: res is nil",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				var res *model.GetEvaluationResponse
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					ErrResponseNil,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: evaluation is nil",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "value")
				res.Evaluation = nil
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					ErrResponseEvaluationNil,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: feature id doesn't match",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, "invalid-feature-id", "value")
				err := fmt.Errorf("%w: actual %s != expected %s",
					ErrResponseFeatureIDMismatch,
					res.Evaluation.FeatureID,
					featureID,
				)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					err,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: variation value is empty",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorEvent(
					ErrResponseVariationValueEmpty,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					res,
					100,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "value",
			isErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual, err := s.getEvaluation(ctx, model.SourceIDGoServer, tt.user, tt.featureID)
			if tt.isErr {
				assert.Error(t, err)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, actual.VariationValue)
			}
		})
	}
}

func TestTrack(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc  string
		setup func(context.Context, *sdk, *user.User, string)
		user  *user.User
	}{
		{
			desc: "invalid user",
			setup: func(ctx context.Context, s *sdk, user *user.User, GoalID string) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGoalEvent(
					user,
					sdkGoalID,
					0.0,
				).Times(0)
			},
			user: newUser(t, ""),
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, GoalID string) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGoalEvent(
					user,
					sdkGoalID,
					0.0,
				)
			},
			user: newUser(t, sdkUserID),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			GoalID := sdkGoalID
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, GoalID)
			}
			s.Track(ctx, tt.user, GoalID)
		})
	}
}

func TestTrackValue(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc  string
		setup func(context.Context, *sdk, *user.User, string, float64)
		user  *user.User
	}{
		{
			desc: "invalid user",
			setup: func(ctx context.Context, s *sdk, user *user.User, GoalID string, value float64) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGoalEvent(
					user,
					sdkGoalID,
					value,
				).Times(0)
			},
			user: newUser(t, ""),
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, GoalID string, value float64) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGoalEvent(
					user,
					sdkGoalID,
					value,
				)
			},
			user: newUser(t, sdkUserID),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			GoalID := sdkGoalID
			value := 1.1
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, GoalID, 1.1)
			}
			s.TrackValue(ctx, tt.user, GoalID, value)
		})
	}
}

func TestClose(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc  string
		setup func(context.Context, *sdk)
		isErr bool
	}{
		{
			desc: "return error when failed to close event processor",
			setup: func(ctx context.Context, s *sdk) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().Drain(ctx).Return(errors.New("error"))
			},
			isErr: true,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().Drain(ctx).Return(nil)
			},
			isErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s)
			}
			err := s.Close(ctx)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEvaluationDetails(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		user         *user.User
		featureID    string
		setup        func(context.Context, *sdk, *user.User, string)
		defaultValue interface{}
		expected     model.BKTEvaluationDetails[interface{}]
	}{
		{
			desc:         "return default value when featureID is empty",
			user:         newUser(t, sdkUserID),
			featureID:    "",
			defaultValue: "default",
			expected: model.NewEvaluationDetails[interface{}](
				"",
				sdkUserID,
				"",
				"",
				0,
				model.ReasonErrorFeatureFlagIDNotSpecified,
				"default",
			),
		},
		{
			desc:         "return default value when user is nil",
			user:         nil,
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected: model.NewEvaluationDetails[interface{}](
				sdkFeatureID,
				"",
				"",
				"",
				0,
				model.ReasonErrorUserIDNotSpecified,
				"default",
			),
		},
		{
			desc:         "return default value when user ID is empty",
			user:         newUser(t, ""),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected: model.NewEvaluationDetails[interface{}](
				sdkFeatureID,
				"",
				"",
				"",
				0,
				model.ReasonErrorUserIDNotSpecified,
				"default",
			),
		},
		{
			desc:      "success validation",
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				res := newGetEvaluationResponse(t, featureID, "true")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					gomock.Any(),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					100,
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					user,
					res.Evaluation,
				)
			},
			defaultValue: "default",
			expected: model.NewEvaluationDetails[interface{}](
				sdkFeatureID,
				sdkUserID,
				"testVersion",
				"testVersionName",
				1,
				model.ReasonTarget,
				"true",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := newSDKWithMock(t, mockCtrl)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(ctx, s, tt.user, tt.featureID)
			}
			actual := getEvaluationDetails(ctx, s, tt.user, tt.featureID, tt.defaultValue, "func")
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestValidateGetEvaluationRequest(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tests := []struct {
		desc         string
		user         *user.User
		featureID    string
		isErr        bool
		defaultValue string
	}{
		{
			desc:         "invalid user",
			user:         newUser(t, ""),
			featureID:    sdkFeatureID,
			isErr:        true,
			defaultValue: "default",
		},
		{
			desc:         "invalid featureId",
			user:         newUser(t, sdkUserID),
			featureID:    "",
			isErr:        true,
			defaultValue: "default",
		},
		{
			desc:         "valid user & featureId",
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			isErr:        false,
			defaultValue: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			s := newSDKWithMock(t, mockCtrl)
			details, ok := validateGetEvaluationRequest(s, tt.user, tt.featureID, tt.defaultValue, "test")
			if tt.isErr {
				assert.False(t, ok)
				assert.Equal(t, tt.defaultValue, details.VariationValue)
				assert.NotEmpty(t, details.Reason)
			} else {
				assert.True(t, ok)
				assert.Equal(t, "", details.VariationValue)
				assert.Empty(t, details.Reason)
			}
		})
	}
}

func newSDKWithMock(t *testing.T, mockCtrl *gomock.Controller) *sdk {
	t.Helper()
	return &sdk{
		tag:            sdkTag,
		version:        sdkVersion,
		sourceID:       model.SourceIDGoServer.Int32(),
		apiClient:      mockapi.NewMockClient(mockCtrl),
		eventProcessor: mockevent.NewMockProcessor(mockCtrl),
		loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
	}
}

func newUser(t *testing.T, id string) *user.User {
	t.Helper()
	return &user.User{ID: id}
}

func TestNewSDK(t *testing.T) {
	t.Parallel()
	tests := []struct {
		desc    string
		opts    []Option
		wantErr error
		isErr   bool
	}{
		{
			desc:    "success",
			opts:    []Option{WithAPIKey("api-key"), WithAPIEndpoint("api.example.com"), WithScheme("https"), WithWrapperSourceID(model.SourceIDGoServer.Int32())},
			wantErr: nil,
			isErr:   false,
		},
		{
			desc:    "empty API key",
			opts:    []Option{WithAPIKey(""), WithAPIEndpoint("api.example.com"), WithScheme("https")},
			wantErr: api.ErrEmptyAPIKey,
			isErr:   true,
		},
		{
			desc:    "invalid scheme",
			opts:    []Option{WithAPIKey("api-key"), WithAPIEndpoint("api.example.com"), WithScheme("invalid")},
			wantErr: api.ErrInvalidScheme,
			isErr:   true,
		},
		{
			desc:    "empty API endpoint",
			opts:    []Option{WithAPIKey("api-key"), WithAPIEndpoint(""), WithScheme("https")},
			wantErr: api.ErrEmptyAPIEndpoint,
			isErr:   true,
		},
		{
			desc:    "invalid sourceID",
			opts:    []Option{WithAPIKey("api-key"), WithAPIEndpoint("api.example.com"), WithScheme("https"), WithWrapperSourceID(123)},
			wantErr: fmt.Errorf("invalid sourceID: 123"),
			isErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			sdk, err := NewSDK(ctx, tt.opts...)
			if tt.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sdk)
			}
		})
	}
}

func newGetEvaluationResponse(t *testing.T, featureID, value string) *model.GetEvaluationResponse {
	t.Helper()
	return &model.GetEvaluationResponse{
		Evaluation: &model.Evaluation{
			FeatureID:      featureID,
			FeatureVersion: 1,
			VariationID:    "testVersion",
			VariationName:  "testVersionName",
			VariationValue: value,
			Reason: &model.Reason{
				RuleID: "test",
				Type:   model.ReasonTarget,
			},
		},
	}
}
