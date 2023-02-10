package bucketeer

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"unsafe"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/api"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/model"
	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
	mockapi "github.com/ca-dp/bucketeer-go-server-sdk/test/mock/api"
	mockevent "github.com/ca-dp/bucketeer-go-server-sdk/test/mock/event"
)

const (
	sdkTag       = "go-server"
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected:     false,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "true")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected:     1.1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected:     "default",
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			dst:       &DstStruct{},
			expected:  &DstStruct{},
		},
		{
			desc: "faled to unmarshal variation",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, `invalid`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user,
					featureID,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "int": "int2"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusGatewayTimeout),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushTimeoutErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusGatewayTimeout)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "get evaluations returns internal error",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					nil,
					api.NewErrStatus(http.StatusInternalServerError),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalSDKErrorMetricsEvent(ctx, model.GetEvaluation)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushErrorStatusCodeMetricsEvent(ctx, model.GetEvaluation, http.StatusInternalServerError)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: res is nil",
			setup: func(ctx context.Context, s *sdk, user *user.User, featureID string) {
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				var res *model.GetEvaluationResponse
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "value")
				res.Evaluation = nil
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, "invalid-feature-id", "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
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
				req := &model.GetEvaluationRequest{
					Tag:       sdkTag,
					User:      user,
					FeatureID: featureID,
					SourceID:  SourceIDGoServer,
				}
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(req).Return(
					res,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					model.GetEvaluation,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(
					ctx,
					int(unsafe.Sizeof(res)),
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
			actual, err := s.getEvaluation(ctx, tt.user, tt.featureID)
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
					ctx,
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
					ctx,
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
					ctx,
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
					ctx,
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
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().Close(ctx).Return(errors.New("error"))
			},
			isErr: true,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk) {
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().Close(ctx).Return(nil)
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
func newSDKWithMock(t *testing.T, mockCtrl *gomock.Controller) *sdk {
	t.Helper()
	return &sdk{
		tag:            sdkTag,
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

func newGetEvaluationResponse(t *testing.T, featureID, value string) *model.GetEvaluationResponse {
	t.Helper()
	return &model.GetEvaluationResponse{
		Evaluation: &model.Evaluation{
			FeatureID:      featureID,
			VariationValue: value,
		},
	}
}
