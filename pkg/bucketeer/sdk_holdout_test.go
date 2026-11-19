package bucketeer

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	evaluation "github.com/bucketeer-io/bucketeer/v2/evaluation/go"
	eventproto "github.com/bucketeer-io/bucketeer/v2/proto/event/client"
	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/v2/proto/gateway"
	userproto "github.com/bucketeer-io/bucketeer/v2/proto/user"

	mockcache "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
	mockevent "github.com/bucketeer-io/go-server-sdk/test/mock/event"
)

const (
	holdoutFeatureID = "feature-id-holdout"
	// The HASH group of newHashHoldoutGroup holds this user out
	holdoutUserIDHeldOut = "user-id-24"
	// The HASH group of newHashHoldoutGroup doesn't hold this user out
	holdoutUserIDNotHeldOut = "user-id-1"
)

// The flag serves the true variation to everyone.
func newHoldoutFeature(t *testing.T, id string) *ftproto.Feature {
	t.Helper()
	return &ftproto.Feature{
		Id:      id,
		Enabled: true,
		Version: 3,
		// The tag of newSDKLocalEvaluationWithMock
		Tags:          []string{"server"},
		UpdatedAt:     1763400000,
		VariationType: ftproto.Feature_BOOLEAN,
		Variations: []*ftproto.Variation{
			{Id: "variation-true-id", Name: "true-name", Value: "true"},
			{Id: "variation-false-id", Name: "false-name", Value: "false"},
		},
		DefaultStrategy: &ftproto.Strategy{
			Type:          ftproto.Strategy_FIXED,
			FixedStrategy: &ftproto.FixedStrategy{Variation: "variation-true-id"},
		},
		OffVariation: "variation-false-id",
	}
}

// The group serves the false variation of the feature to the users it holds out.
// It has only the fields the server sends to the SDK.
func newHashHoldoutGroup(t *testing.T, id, featureID string) *ftproto.HoldoutGroup {
	t.Helper()
	group := &ftproto.HoldoutGroup{
		Id:             id,
		EnvironmentId:  "environment-id",
		AssignmentType: ftproto.HoldoutGroup_HASH,
		SamplingSeed:   "sampling-seed",
		// It wraps: [97, 100) and [0, 2)
		BucketStart: 97,
		Percentage:  5,
		Status:      ftproto.HoldoutGroup_RUNNING,
		UpdatedAt:   1763500000,
		Features: []*ftproto.HoldoutGroupFeature{
			{
				Id:                 id + ":" + featureID,
				HoldoutGroupId:     id,
				FeatureId:          featureID,
				EnvironmentId:      "environment-id",
				HoldoutVariationId: "variation-false-id",
				AddedAt:            1763400000,
			},
		},
	}
	require.True(t, evaluation.IsHeldOut(group, &userproto.User{Id: holdoutUserIDHeldOut}, nil))
	require.False(t, evaluation.IsHeldOut(group, &userproto.User{Id: holdoutUserIDNotHeldOut}, nil))
	return group
}

func newSegmentHoldoutGroup(t *testing.T, id, featureID, segmentID string) *ftproto.HoldoutGroup {
	t.Helper()
	return &ftproto.HoldoutGroup{
		Id:             id,
		EnvironmentId:  "environment-id",
		AssignmentType: ftproto.HoldoutGroup_SEGMENT,
		SegmentId:      segmentID,
		Status:         ftproto.HoldoutGroup_RUNNING,
		UpdatedAt:      1763500001,
		Features: []*ftproto.HoldoutGroupFeature{
			{
				Id:                 id + ":" + featureID,
				HoldoutGroupId:     id,
				FeatureId:          featureID,
				EnvironmentId:      "environment-id",
				HoldoutVariationId: "variation-false-id",
				AddedAt:            1763400001,
			},
		},
	}
}

func TestLocalBoolVariationDetailsHoldout(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	feature := newHoldoutFeature(t, holdoutFeatureID)
	group := newHashHoldoutGroup(t, "holdout-group-id", holdoutFeatureID)

	patterns := []struct {
		desc               string
		userID             string
		defaultValue       bool
		expectedEvaluation *model.Evaluation
		expected           model.BKTEvaluationDetails[bool]
	}{
		{
			desc:         "the held-out user gets the holdout variation, and the event has the reason and the group ID",
			userID:       holdoutUserIDHeldOut,
			defaultValue: true,
			expectedEvaluation: &model.Evaluation{
				ID:             holdoutFeatureID + ":3:" + holdoutUserIDHeldOut,
				FeatureID:      holdoutFeatureID,
				FeatureVersion: 3,
				UserID:         holdoutUserIDHeldOut,
				VariationID:    "variation-false-id",
				VariationName:  "false-name",
				VariationValue: "false",
				Reason:         &model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: "holdout-group-id"},
			},
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      holdoutFeatureID,
				FeatureVersion: 3,
				UserID:         holdoutUserIDHeldOut,
				VariationID:    "variation-false-id",
				VariationName:  "false-name",
				VariationValue: false,
				Reason:         model.EvaluationReasonHoldout,
			},
		},
		{
			desc:         "the user who is not held out gets the variation of the flag, and the event has no group ID",
			userID:       holdoutUserIDNotHeldOut,
			defaultValue: false,
			expectedEvaluation: &model.Evaluation{
				ID:             holdoutFeatureID + ":3:" + holdoutUserIDNotHeldOut,
				FeatureID:      holdoutFeatureID,
				FeatureVersion: 3,
				UserID:         holdoutUserIDNotHeldOut,
				VariationID:    "variation-true-id",
				VariationName:  "true-name",
				VariationValue: "true",
				Reason:         &model.Reason{Type: model.ReasonDefault},
			},
			expected: model.BKTEvaluationDetails[bool]{
				FeatureID:      holdoutFeatureID,
				FeatureVersion: 3,
				UserID:         holdoutUserIDNotHeldOut,
				VariationID:    "variation-true-id",
				VariationName:  "true-name",
				VariationValue: true,
				Reason:         model.EvaluationReasonDefault,
			},
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			s := newSDKLocalEvaluationWithMock(t, controller)
			require.NoError(t, s.holdoutGroupsCache.Put([]*ftproto.HoldoutGroup{group}))
			u := &user.User{ID: p.userID}
			s.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Get(holdoutFeatureID).Return(feature, nil)
			s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(
				gomock.Any(), model.SDKGetEvaluation,
			)
			s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(u, p.expectedEvaluation)

			actual := s.BoolVariationDetails(context.Background(), u, holdoutFeatureID, p.defaultValue)
			assert.Equal(t, p.expected, actual)
		})
	}
}

// The server evaluates the holdout group in the remote mode.
// The SDK must pass the reason and the group ID it returns on to the evaluation event.
func TestBoolVariationDetailsHoldout(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	s := newSDKWithMock(t, controller)
	u := newUser(t, sdkUserID)
	res := &model.GetEvaluationResponse{
		Evaluation: &model.Evaluation{
			ID:             sdkFeatureID + ":3:" + sdkUserID,
			FeatureID:      sdkFeatureID,
			FeatureVersion: 3,
			UserID:         sdkUserID,
			VariationID:    "variation-false-id",
			VariationName:  "false-name",
			VariationValue: "false",
			Reason:         &model.Reason{Type: model.ReasonHoldout, HoldoutGroupID: "holdout-group-id"},
		},
	}
	s.apiClient.(*mockapi.MockClient).EXPECT().GetEvaluation(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, 100, nil)
	s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetEvaluation)
	s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushSizeMetricsEvent(100, model.GetEvaluation)
	s.eventProcessor.(*mockevent.MockProcessor).EXPECT().PushEvaluationEvent(
		u,
		gomock.Cond(func(x any) bool {
			e, ok := x.(*model.Evaluation)
			return ok && e.Reason != nil &&
				e.Reason.Type == model.ReasonHoldout &&
				e.Reason.HoldoutGroupID == "holdout-group-id"
		}),
	)

	actual := s.BoolVariationDetails(context.Background(), u, sdkFeatureID, true)
	assert.Equal(t, model.BKTEvaluationDetails[bool]{
		FeatureID:      sdkFeatureID,
		FeatureVersion: 3,
		UserID:         sdkUserID,
		VariationID:    "variation-false-id",
		VariationName:  "false-name",
		VariationValue: false,
		Reason:         model.EvaluationReasonHoldout,
	}, actual)
}

// fakeGateway answers the SDK as the API gateway does. It encodes the proto messages of the
// server with protojson, and decodes the requests of the SDK into them, so the tests exercise
// the JSON the SDK and the server exchange. It doesn't evaluate anything.
type fakeGateway struct {
	t  *testing.T
	mu sync.Mutex

	features []*ftproto.Feature
	// The holdout groups that are serving
	groups []*ftproto.HoldoutGroup
	// All the segments of the environment, by ID
	segments map[string]*ftproto.SegmentUsers
	// What get_evaluation returns, by user ID
	evaluations map[string]*ftproto.Evaluation

	// The featureFlagsIDs returned to the SDK
	issuedIDs map[string]struct{}
	// The featureFlagsIDs sent by the SDK
	requestedIDs []string
	// The evaluation events registered by the SDK
	evaluationEvents []*eventproto.EvaluationEvent
}

var (
	// See pkg/rpc/gateway/gateway.go of the bucketeer repository
	fakeGatewayMarshalOptions   = protojson.MarshalOptions{EmitUnpopulated: true}
	fakeGatewayUnmarshalOptions = protojson.UnmarshalOptions{DiscardUnknown: true}
)

func newFakeGateway(t *testing.T) (*fakeGateway, string) {
	t.Helper()
	g := &fakeGateway{
		t:           t,
		segments:    map[string]*ftproto.SegmentUsers{},
		evaluations: map[string]*ftproto.Evaluation{},
		issuedIDs:   map[string]struct{}{},
	}
	server := httptest.NewServer(g)
	t.Cleanup(server.Close)
	return g, strings.TrimPrefix(server.URL, "http://")
}

func (g *fakeGateway) setHoldoutGroups(groups ...*ftproto.HoldoutGroup) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.groups = groups
}

func (g *fakeGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		g.t.Errorf("failed to read the request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	var req, resp proto.Message
	var handle func()
	switch r.URL.Path {
	case "/get_feature_flags":
		in := &gwproto.GetFeatureFlagsRequest{}
		req, handle = in, func() { resp = g.getFeatureFlags(in) }
	case "/get_segment_users":
		in := &gwproto.GetSegmentUsersRequest{}
		req, handle = in, func() { resp = g.getSegmentUsers(in) }
	case "/get_evaluation":
		in := &gwproto.GetEvaluationRequest{}
		req, handle = in, func() { resp = &gwproto.GetEvaluationResponse{Evaluation: g.evaluations[in.User.GetId()]} }
	case "/register_events":
		in := &gwproto.RegisterEventsRequest{}
		req, handle = in, func() { resp = g.registerEvents(in) }
	default:
		g.t.Errorf("unexpected request: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := fakeGatewayUnmarshalOptions.Unmarshal(body, req); err != nil {
		g.t.Errorf("failed to decode the request of %s: %v: %s", r.URL.Path, err, body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	handle()
	encoded, err := fakeGatewayMarshalOptions.Marshal(resp)
	if err != nil {
		g.t.Errorf("failed to encode the response of %s: %v", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(encoded)
}

// See GetFeatureFlags in pkg/api/api/api_grpc.go of the bucketeer repository.
func (g *fakeGateway) getFeatureFlags(req *gwproto.GetFeatureFlagsRequest) *gwproto.GetFeatureFlagsResponse {
	g.requestedIDs = append(g.requestedIDs, req.FeatureFlagsId)
	// The ID has a second part while a group is serving: "<flags>:<holdout>"
	id := evaluation.GenerateFeaturesID(g.features, g.groups...)
	g.issuedIDs[id] = struct{}{}
	now := time.Now().Unix()
	// Nothing changed, or only the holdout groups changed: an empty flag diff with the groups
	if req.FeatureFlagsId == id ||
		(req.FeatureFlagsId != "" &&
			evaluation.FeaturesIDFlagsPart(req.FeatureFlagsId) == evaluation.FeaturesIDFlagsPart(id)) {
		return &gwproto.GetFeatureFlagsResponse{
			FeatureFlagsId:         id,
			Features:               []*ftproto.Feature{},
			ArchivedFeatureFlagIds: []string{},
			RequestedAt:            min(req.RequestedAt, now),
			HoldoutGroups:          g.groups,
		}
	}
	return &gwproto.GetFeatureFlagsResponse{
		FeatureFlagsId:         id,
		Features:               g.features,
		ArchivedFeatureFlagIds: []string{},
		RequestedAt:            now,
		ForceUpdate:            true,
		HoldoutGroups:          g.groups,
	}
}

// See GetSegmentUsers in pkg/api/api/api_grpc.go of the bucketeer repository.
// No flag rule references a segment here, so the segments of the holdout groups are the only targets.
func (g *fakeGateway) getSegmentUsers(req *gwproto.GetSegmentUsersRequest) *gwproto.GetSegmentUsersResponse {
	now := time.Now().Unix()
	targetIDs := evaluation.HoldoutSegmentIDs(g.groups)
	if len(targetIDs) == 0 {
		return &gwproto.GetSegmentUsersResponse{
			SegmentUsers:      []*ftproto.SegmentUsers{},
			DeletedSegmentIds: []string{},
			RequestedAt:       now,
			ForceUpdate:       true,
		}
	}
	contains := func(ids []string, id string) bool {
		for _, v := range ids {
			if v == id {
				return true
			}
		}
		return false
	}
	deleted := []string{}
	for _, id := range req.SegmentIds {
		if !contains(targetIDs, id) {
			deleted = append(deleted, id)
		}
	}
	// The segment of a holdout group is sent when the SDK doesn't report it, whatever its age
	updated := []*ftproto.SegmentUsers{}
	for _, id := range targetIDs {
		segment := g.segments[id]
		if segment.UpdatedAt >= req.RequestedAt || !contains(req.SegmentIds, id) {
			updated = append(updated, segment)
		}
	}
	if len(updated) == 0 && len(deleted) == 0 {
		now = req.RequestedAt
	}
	return &gwproto.GetSegmentUsersResponse{
		SegmentUsers:      updated,
		DeletedSegmentIds: deleted,
		RequestedAt:       now,
	}
}

func (g *fakeGateway) registerEvents(req *gwproto.RegisterEventsRequest) *gwproto.RegisterEventsResponse {
	for _, event := range req.Events {
		evaluationEvent := &eventproto.EvaluationEvent{}
		if !event.Event.MessageIs(evaluationEvent) {
			continue
		}
		if err := event.Event.UnmarshalTo(evaluationEvent); err != nil {
			g.t.Errorf("failed to decode the evaluation event: %v", err)
			continue
		}
		g.evaluationEvents = append(g.evaluationEvents, evaluationEvent)
	}
	return &gwproto.RegisterEventsResponse{}
}

// lastEvaluationEvent returns the last evaluation event registered for the user and the reason.
func (g *fakeGateway) lastEvaluationEvent(userID string, reason ftproto.Reason_Type) *eventproto.EvaluationEvent {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i := len(g.evaluationEvents) - 1; i >= 0; i-- {
		event := g.evaluationEvents[i]
		if event.User.GetId() == userID && event.Reason.GetType() == reason {
			return event
		}
	}
	return nil
}

// The SDK polls the fake gateway, evaluates locally, and registers the evaluation events.
func TestLocalEvaluationHoldoutWithGateway(t *testing.T) {
	t.Parallel()
	const (
		otherFeatureID = "feature-id-other"
		hashGroupID    = "holdout-group-id-hash"
		segmentGroupID = "holdout-group-id-segment"
		segmentID      = "segment-id-holdout"
		segmentUserID  = "user-id-in-segment"
	)
	gateway, endpoint := newFakeGateway(t)
	gateway.features = []*ftproto.Feature{
		newHoldoutFeature(t, holdoutFeatureID),
		newHoldoutFeature(t, otherFeatureID),
	}
	// The segment exists before a group references it, and it is older than any request of the SDK
	gateway.segments[segmentID] = &ftproto.SegmentUsers{
		SegmentId: segmentID,
		UpdatedAt: 1763000000,
		Users: []*ftproto.SegmentUser{
			{
				Id:        segmentID + ":" + segmentUserID,
				SegmentId: segmentID,
				UserId:    segmentUserID,
				State:     ftproto.SegmentUser_INCLUDED,
			},
		},
	}
	gateway.setHoldoutGroups(newHashHoldoutGroup(t, hashGroupID, holdoutFeatureID))

	ctx := context.Background()
	sdk, err := NewSDK(
		ctx,
		WithAPIKey("api-key"),
		WithAPIEndpoint(endpoint),
		WithScheme("http"),
		WithEnableLocalEvaluation(true),
		WithCachePollingInterval(20*time.Millisecond),
		WithErrorLogger(log.DiscardErrorLogger),
	)
	require.NoError(t, err)

	type result struct {
		value  bool
		reason model.EvaluationReason
	}
	// The default value is the opposite of the expected one, so it can't pass for it
	evaluate := func(userID, featureID string, expected bool) result {
		details := sdk.BoolVariationDetails(ctx, user.NewUser(userID, nil), featureID, !expected)
		return result{value: details.VariationValue, reason: details.Reason}
	}
	eventually := func(desc string, userID, featureID string, expected result) {
		t.Helper()
		var actual result
		require.Eventuallyf(t, func() bool {
			actual = evaluate(userID, featureID, expected.value)
			return actual == expected
		}, 10*time.Second, 5*time.Millisecond, "%s: expected %+v, the last result was %+v", desc, expected, &actual)
	}
	holdout := result{value: false, reason: model.EvaluationReasonHoldout}
	normal := result{value: true, reason: model.EvaluationReasonDefault}

	// The HASH group is serving
	eventually("hash group: the held-out user", holdoutUserIDHeldOut, holdoutFeatureID, holdout)
	assert.Equal(t, normal, evaluate(holdoutUserIDNotHeldOut, holdoutFeatureID, true))
	assert.Equal(t, normal, evaluate(holdoutUserIDHeldOut, otherFeatureID, true))

	// The group stopped. It reaches the SDK with an empty flag diff and an empty list
	gateway.setHoldoutGroups()
	eventually("no group: the user who was held out", holdoutUserIDHeldOut, holdoutFeatureID, normal)

	// A SEGMENT group started. Its segment reaches the SDK even though no flag rule references it
	gateway.setHoldoutGroups(newSegmentHoldoutGroup(t, segmentGroupID, holdoutFeatureID, segmentID))
	eventually("segment group: the user in the segment", segmentUserID, holdoutFeatureID, holdout)
	assert.Equal(t, normal, evaluate(holdoutUserIDHeldOut, holdoutFeatureID, true))
	assert.Equal(t, normal, evaluate(segmentUserID, otherFeatureID, true))

	// The group stopped
	gateway.setHoldoutGroups()
	eventually("no group: the user in the segment", segmentUserID, holdoutFeatureID, normal)

	// The events are sent when the SDK is closed
	require.NoError(t, sdk.Close(ctx))

	// The evaluation events have the reason and the ID of the group that served the variation
	event := gateway.lastEvaluationEvent(holdoutUserIDHeldOut, ftproto.Reason_HOLDOUT)
	if assert.NotNil(t, event) {
		assert.Equal(t, hashGroupID, event.Reason.HoldoutGroupId)
		assert.Equal(t, holdoutFeatureID, event.FeatureId)
		assert.Equal(t, int32(3), event.FeatureVersion)
		assert.Equal(t, "variation-false-id", event.VariationId)
		assert.Equal(t, eventproto.SourceId_GO_SERVER, event.SourceId)
	}
	event = gateway.lastEvaluationEvent(segmentUserID, ftproto.Reason_HOLDOUT)
	if assert.NotNil(t, event) {
		assert.Equal(t, segmentGroupID, event.Reason.HoldoutGroupId)
		assert.Equal(t, holdoutFeatureID, event.FeatureId)
		assert.Equal(t, "variation-false-id", event.VariationId)
	}
	gateway.mu.Lock()
	defer gateway.mu.Unlock()
	for _, event := range gateway.evaluationEvents {
		if event.Reason.GetType() != ftproto.Reason_HOLDOUT {
			assert.Empty(t, event.Reason.GetHoldoutGroupId(), "reason: %s", event.Reason.GetType())
			continue
		}
		assert.Equal(t, holdoutFeatureID, event.FeatureId)
	}

	// The SDK sends back the featureFlagsID as it received it, including the one with the holdout part
	sentIDWithHoldoutPart := false
	for _, id := range gateway.requestedIDs {
		if id == "" {
			continue
		}
		assert.Contains(t, gateway.issuedIDs, id)
		if strings.Contains(id, ":") {
			sentIDWithHoldoutPart = true
		}
	}
	assert.True(t, sentIDWithHoldoutPart)
}

// The server evaluates the user. The SDK gets the evaluation with the reason and the group ID,
// and registers the evaluation event with them.
func TestRemoteEvaluationHoldoutWithGateway(t *testing.T) {
	t.Parallel()
	gateway, endpoint := newFakeGateway(t)
	newEvaluation := func(userID, variationID, value string, reason *ftproto.Reason) *ftproto.Evaluation {
		return &ftproto.Evaluation{
			Id:             holdoutFeatureID + ":3:" + userID,
			FeatureId:      holdoutFeatureID,
			FeatureVersion: 3,
			UserId:         userID,
			VariationId:    variationID,
			VariationName:  value + "-name",
			VariationValue: value,
			Reason:         reason,
		}
	}
	gateway.evaluations[holdoutUserIDHeldOut] = newEvaluation(
		holdoutUserIDHeldOut, "variation-false-id", "false",
		&ftproto.Reason{Type: ftproto.Reason_HOLDOUT, HoldoutGroupId: "holdout-group-id"},
	)
	gateway.evaluations[holdoutUserIDNotHeldOut] = newEvaluation(
		holdoutUserIDNotHeldOut, "variation-true-id", "true",
		&ftproto.Reason{Type: ftproto.Reason_DEFAULT},
	)

	ctx := context.Background()
	sdk, err := NewSDK(
		ctx,
		WithAPIKey("api-key"),
		WithAPIEndpoint(endpoint),
		WithScheme("http"),
		WithErrorLogger(log.DiscardErrorLogger),
	)
	require.NoError(t, err)

	details := sdk.BoolVariationDetails(ctx, user.NewUser(holdoutUserIDHeldOut, nil), holdoutFeatureID, true)
	assert.False(t, details.VariationValue)
	assert.Equal(t, "variation-false-id", details.VariationID)
	assert.Equal(t, model.EvaluationReasonHoldout, details.Reason)

	details = sdk.BoolVariationDetails(ctx, user.NewUser(holdoutUserIDNotHeldOut, nil), holdoutFeatureID, false)
	assert.True(t, details.VariationValue)
	assert.Equal(t, model.EvaluationReasonDefault, details.Reason)

	// The events are sent when the SDK is closed
	require.NoError(t, sdk.Close(ctx))

	event := gateway.lastEvaluationEvent(holdoutUserIDHeldOut, ftproto.Reason_HOLDOUT)
	if assert.NotNil(t, event) {
		assert.Equal(t, "holdout-group-id", event.Reason.HoldoutGroupId)
		assert.Equal(t, holdoutFeatureID, event.FeatureId)
		assert.Equal(t, int32(3), event.FeatureVersion)
		assert.Equal(t, "variation-false-id", event.VariationId)
	}
	event = gateway.lastEvaluationEvent(holdoutUserIDNotHeldOut, ftproto.Reason_DEFAULT)
	if assert.NotNil(t, event) {
		assert.Empty(t, event.Reason.HoldoutGroupId)
		assert.Equal(t, "variation-true-id", event.VariationId)
	}
}
