package bucketeer

import (
	"context"
	"testing"

	"github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/log"
	protofeature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	protogateway "github.com/ca-dp/bucketeer-go-server-sdk/proto/gateway"
	mockapi "github.com/ca-dp/bucketeer-go-server-sdk/test/mock/api"
	mockevent "github.com/ca-dp/bucketeer-go-server-sdk/test/mock/event"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		setup        func(context.Context, *sdk, *User, string)
		user         *User
		featureID    string
		defaultValue bool
		expected     bool
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
					featureID,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: false,
			expected:     false,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
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
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "true")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
					user.User,
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
		setup        func(context.Context, *sdk, *User, string)
		user         *User
		featureID    string
		defaultValue int
		expected     int
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
					featureID,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
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
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
					user.User,
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
		setup        func(context.Context, *sdk, *User, string)
		user         *User
		featureID    string
		defaultValue int64
		expected     int64
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
					featureID,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1,
			expected:     1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
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
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
					user.User,
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
		setup        func(context.Context, *sdk, *User, string)
		user         *User
		featureID    string
		defaultValue float64
		expected     float64
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
					featureID,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: 1.1,
			expected:     1.1,
		},
		{
			desc: "return default value when faled to parse variation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "invalid")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
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
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "2.2")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
					user.User,
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
		setup        func(context.Context, *sdk, *User, string)
		user         *User
		featureID    string
		defaultValue string
		expected     string
	}{
		{
			desc: "return default value when failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
					featureID,
				)
			},
			user:         newUser(t, sdkUserID),
			featureID:    sdkFeatureID,
			defaultValue: "default",
			expected:     "default",
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
					user.User,
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
		setup     func(context.Context, *sdk, *User, string)
		user      *User
		featureID string
		dst       interface{}
		expected  interface{}
	}{
		{
			desc: "failed to get evaluation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
					featureID,
				)
			},
			user:      newUser(t, sdkUserID),
			featureID: sdkFeatureID,
			dst:       &DstStruct{},
			expected:  &DstStruct{},
		},
		{
			desc: "faled to unmarshal variation",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, `invalid`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushDefaultEvaluationEvent(
					ctx,
					user.User,
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
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, `{"str": "str2", "int": "int2"}`)
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
					ctx,
					user.User,
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
		setup         func(context.Context, *sdk, *User, string)
		user          *User
		featureID     string
		expectedValue string
		isErr         bool
	}{
		{
			desc: "get evaluations returns timeout error",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.DeadlineExceeded, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushTimeoutErrorCountMetricsEvent(ctx, sdkTag)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "get evaluations returns internal error",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					nil,
					status.Error(codes.Internal, "error"),
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushInternalErrorCountMetricsEvent(ctx, sdkTag)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: evaluation is nil",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "value")
				res.Evaluation = nil
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: feature id doesn't match",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, "invalid-feature-id", "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "invalid get evaluation res: variation is nil",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "value")
				res.Evaluation.Variation = nil
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(res, nil)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
				)
			},
			user:          newUser(t, sdkUserID),
			featureID:     sdkFeatureID,
			expectedValue: "",
			isErr:         true,
		},
		{
			desc: "success",
			setup: func(ctx context.Context, s *sdk, user *User, featureID string) {
				req := &protogateway.GetEvaluationRequest{Tag: sdkTag, User: user.User, FeatureId: featureID}
				res := newGetEvaluationResponse(t, featureID, "value")
				s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(ctx, req).Return(
					res,
					nil,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationLatencyMetricsEvent(
					ctx,
					gomock.Any(), // duration
					sdkTag,
				)
				s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGetEvaluationSizeMetricsEvent(
					ctx,
					proto.Size(res),
					sdkTag,
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
				assert.Equal(t, tt.expectedValue, actual.Variation.Value)
			}
		})
	}
}

func TestTrack(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := context.Background()
	user := newUser(t, sdkUserID)
	s := newSDKWithMock(t, mockCtrl)
	s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGoalEvent(ctx, user.User, sdkGoalID, 0.0)
	s.Track(ctx, user, sdkGoalID)
}

func TestTrackValue(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := context.Background()
	user := newUser(t, sdkUserID)
	value := 1.1
	s := newSDKWithMock(t, mockCtrl)
	s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushGoalEvent(ctx, user.User, sdkGoalID, value)
	s.TrackValue(ctx, user, sdkGoalID, value)
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

func newUser(t *testing.T, id string) *User {
	t.Helper()
	u, err := NewUser(id, nil)
	require.NoError(t, err)
	return u
}

func newGetEvaluationResponse(t *testing.T, featureID, value string) *protogateway.GetEvaluationResponse {
	t.Helper()
	return &protogateway.GetEvaluationResponse{
		Evaluation: &protofeature.Evaluation{
			FeatureId: featureID,
			Variation: &protofeature.Variation{Value: value},
		},
	}
}
