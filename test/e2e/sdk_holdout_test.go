package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

// Holdout group tests
//
// A holdout group is not a permanent fixture: it has a stop date, only one group can serve
// per environment, and it can only be created with the web API. So these tests are skipped
// unless -holdout-feature-id is passed, and they are not part of `make e2e`.
//
// They require the following in the test environment (Bucketeer server 2.4.0 or later):
//
// Flag passed with -holdout-feature-id (string):
//   - Enabled, with the tag "go-server"
//   - A fixed default strategy, and no targeting rule
//
// Holdout group in the RUNNING status:
//   - Targets the flag with a holdout variation that is not the variation of the default strategy
//   - HASH: a percentage large enough to hold out some of the holdoutUsers users (20% is plenty)
//   - SEGMENT: a segment that no flag rule references. Pass one of its users with
//     -holdout-segment-user-id
//
// Run them with the following. The API endpoint is the host only, without the scheme:
//
//	go test -v -race -count=1 ./test/e2e/ -run 'TestHoldout' -args \
//	  -api-key=${API_KEY} -api-key-server=${API_KEY_SERVER} -api-endpoint=${API_ENDPOINT} \
//	  -holdout-feature-id=<flag ID>
//
// After the group is stopped, pass -holdout-serving=false to verify that nobody is held out.
//
// The tests verify that the server accepts the evaluation events, not what it stores.
// Check on the backend that the stored events have the HOLDOUT reason and the holdout group ID.
var (
	holdoutFeatureID = flag.String(
		"holdout-feature-id", "",
		"ID of a flag targeted by the holdout group. The holdout tests are skipped when it is empty",
	)
	holdoutSegmentUserID = flag.String(
		"holdout-segment-user-id", "",
		"ID of a user in the segment of the holdout group, for a group with the SEGMENT assignment",
	)
	holdoutServing = flag.Bool(
		"holdout-serving", true,
		"false when no holdout group targets the flag, to verify that nobody is held out",
	)
)

const (
	// The number of users evaluated locally and by the server
	holdoutUsers = 300
	// A holdout change reaches the gateway within seconds.
	// The worst case is a missed cache refresh and an expired in-memory cache.
	holdoutPropagationTimeout = 150 * time.Second
	holdoutServerAttempts     = 3
)

// holdoutTarget is the holdout group that targets the flag, as the gateway delivers it.
type holdoutTarget struct {
	group   model.HoldoutGroup
	feature model.HoldoutGroupFeature
}

// findHoldoutTarget returns the holdout group that targets the flag in the GetFeatureFlags response.
// It reads the JSON the gateway really sends, which the unit tests can only imitate.
//
// It makes no assertion, because it is called while polling.
func findHoldoutTarget(t *testing.T, featureID string) (*holdoutTarget, error) {
	t.Helper()
	client := newAPIClient(t, *apiKeyServer)
	resp, _, err := client.GetFeatureFlags(
		context.Background(),
		model.NewGetFeatureFlagsRequest(tag, "", sdkVersion, sourceID, 0),
		time.Now().Add(30*time.Second),
	)
	if err != nil {
		return nil, err
	}
	if !findFeature(t, resp.Features, featureID) {
		return nil, fmt.Errorf("the flag %s is not delivered. It must be enabled and have the tag %s", featureID, tag)
	}
	for _, group := range resp.HoldoutGroups {
		for _, feature := range group.Features {
			if feature.FeatureID == featureID {
				return &holdoutTarget{group: group, feature: feature}, nil
			}
		}
	}
	return nil, nil
}

// rawHoldoutGroups returns the holdoutGroups field of the GetFeatureFlags response as the gateway sends it,
// before the SDK decodes it. It is only used to explain a failure.
func rawHoldoutGroups(t *testing.T) string {
	t.Helper()
	body, err := json.Marshal(model.NewGetFeatureFlagsRequest(tag, "", sdkVersion, sourceID, 0))
	if err != nil {
		return err.Error()
	}
	url := fmt.Sprintf("%s://%s/get_feature_flags", *scheme, *apiEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err.Error()
	}
	req.Header.Add("authorization", *apiKeyServer)
	req.Header.Add("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("status %d: %s", resp.StatusCode, respBody)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(respBody, &fields); err != nil {
		return fmt.Sprintf("the response is not a JSON object: %v", err)
	}
	groups, ok := fields["holdoutGroups"]
	if !ok {
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return fmt.Sprintf("the response has no holdoutGroups field. Its fields: %v", keys)
	}
	return string(groups)
}

// requireHoldoutTarget checks the fields the local evaluation depends on.
func requireHoldoutTarget(t *testing.T, target *holdoutTarget) {
	t.Helper()
	group, feature := target.group, target.feature
	require.NotEmpty(t, group.ID)
	require.NotEmpty(t, group.EnvironmentID)
	require.Contains(t, []string{"RUNNING", "ANALYSIS"}, group.Status)
	require.Equal(t, group.ID, feature.HoldoutGroupID)
	require.NotEmpty(t, feature.HoldoutVariationID)
	switch group.AssignmentType {
	case "HASH":
		require.NotEmpty(t, group.SamplingSeed)
		require.Greater(t, group.Percentage, float64(0))
	case "SEGMENT":
		require.NotEmpty(t, group.SegmentID)
	default:
		require.Failf(t, "unknown assignment type", "assignmentType: %q", group.AssignmentType)
	}
}

// getServerEvaluation returns the evaluation of the server, which is what the local evaluation must match.
// The shared backend sometimes answers slowly, so a request that fails is sent again,
// and every failure is logged so that a slow backend isn't mistaken for a holdout problem.
func getServerEvaluation(t *testing.T, client api.Client, featureID string, u *user.User) *model.Evaluation {
	t.Helper()
	var err error
	for attempt := 1; attempt <= holdoutServerAttempts; attempt++ {
		var res *model.GetEvaluationResponse
		res, _, err = client.GetEvaluation(
			context.Background(),
			model.NewGetEvaluationRequest(tag, featureID, sdkVersion, sourceID, u),
			time.Now().Add(30*time.Second),
		)
		if err == nil {
			require.NotNil(t, res.Evaluation)
			require.NotNil(t, res.Evaluation.Reason)
			return res.Evaluation
		}
		t.Logf("GetEvaluation failed (attempt: %d/%d, userID: %s): %v", attempt, holdoutServerAttempts, u.ID, err)
		time.Sleep(5 * time.Second)
	}
	require.NoError(t, err, "the server didn't evaluate %s. This is a request failure, not a holdout result", u.ID)
	return nil
}

func newHoldoutUsers(t *testing.T) []*user.User {
	t.Helper()
	users := make([]*user.User, 0, holdoutUsers+1)
	if *holdoutSegmentUserID != "" {
		users = append(users, newUser(t, *holdoutSegmentUserID))
	}
	for i := 0; i < holdoutUsers; i++ {
		users = append(users, newUser(t, fmt.Sprintf("bucketeer-go-server-holdout-user-id-%d", i)))
	}
	return users
}

// newHoldoutLocalSDK creates an SDK that evaluates locally.
// Its event queue holds the events of all the users, which newLocalSDK would drop,
// and it doesn't output the debug logs, which print every response.
func newHoldoutLocalSDK(t *testing.T, ctx context.Context) bucketeer.SDK {
	t.Helper()
	sdk, err := bucketeer.NewSDK(
		ctx,
		bucketeer.WithCachePollingInterval(5*time.Second),
		bucketeer.WithEnableLocalEvaluation(true),
		bucketeer.WithTag(tag),
		bucketeer.WithAPIKey(*apiKeyServer),
		bucketeer.WithAPIEndpoint(*apiEndpoint),
		bucketeer.WithScheme(*scheme),
		bucketeer.WithEventQueueCapacity(100*holdoutUsers),
		bucketeer.WithWrapperSourceID(sourceID),
		bucketeer.WithWrapperSDKVersion(sdkVersion),
	)
	require.NoError(t, err)
	return sdk
}

// TestHoldoutLocalEvaluationMatchesServer evaluates the same users locally and on the server.
// The server decides who is held out, so the test doesn't need to know the sampling seed of the group.
func TestHoldoutLocalEvaluationMatchesServer(t *testing.T) {
	if *holdoutFeatureID == "" {
		t.Skip("-holdout-feature-id is not set")
	}
	featureID := *holdoutFeatureID
	ctx := context.Background()

	// The group reaches the SDK in the GetFeatureFlags response
	var target *holdoutTarget
	var targetErr error
	delivered := assert.Eventually(t, func() bool {
		target, targetErr = findHoldoutTarget(t, featureID)
		return targetErr == nil && (target != nil) == *holdoutServing
	}, holdoutPropagationTimeout, 2*time.Second,
		"GetFeatureFlags must deliver a holdout group that targets %s: %t", featureID, *holdoutServing)
	if !delivered {
		// Tell a request that fails from a response without the group,
		// and a group the server doesn't send from a group the SDK doesn't decode
		t.Logf("the last error of GetFeatureFlags: %v", targetErr)
		t.Logf("the holdoutGroups the gateway sends: %s", rawHoldoutGroups(t))
		t.FailNow()
	}
	if *holdoutServing {
		requireHoldoutTarget(t, target)
		t.Logf("holdout group: %+v", target.group)
	}

	// The evaluations of the server
	users := newHoldoutUsers(t)
	expected := make(map[string]*model.Evaluation, len(users))
	var heldOut, notHeldOut []*user.User
	client := newAPIClient(t, *apiKey)
	serverStart := time.Now()
	for _, u := range users {
		evaluation := getServerEvaluation(t, client, featureID, u)
		expected[u.ID] = evaluation
		if evaluation.Reason.Type == model.ReasonHoldout {
			heldOut = append(heldOut, u)
			continue
		}
		notHeldOut = append(notHeldOut, u)
	}
	t.Logf("server: %d held-out users, %d other users (%s)",
		len(heldOut), len(notHeldOut), time.Since(serverStart).Round(time.Second))

	if *holdoutServing {
		// A run in which the server holds nobody out, or everybody, proves nothing
		require.NotEmpty(t, heldOut, "the server holds no user out")
		require.NotEmpty(t, notHeldOut, "the server holds every user out")
		for _, u := range heldOut {
			evaluation := expected[u.ID]
			require.Equal(t, target.group.ID, evaluation.Reason.HoldoutGroupID, "userID: %s", u.ID)
			require.Equal(t, target.feature.HoldoutVariationID, evaluation.VariationID, "userID: %s", u.ID)
		}
		if *holdoutSegmentUserID != "" {
			require.Equal(t, "SEGMENT", target.group.AssignmentType)
			require.Equal(t, model.ReasonHoldout, expected[*holdoutSegmentUserID].Reason.Type,
				"the server doesn't hold the user of the segment out: %s", *holdoutSegmentUserID)
		}
	} else {
		require.Empty(t, heldOut, "the server still holds users out")
	}

	// The local evaluation
	sdk := newHoldoutLocalSDK(t, ctx)
	waitForLocalCacheReady(ctx, t, sdk, users[0], featureID)
	if *holdoutServing {
		// The SDK polls the flags and the segment users separately,
		// so wait until it holds the group, and the segment of a SEGMENT group
		require.Eventually(t, func() bool {
			details := sdk.StringVariationDetails(ctx, heldOut[0], featureID, "default")
			return details.Reason == model.EvaluationReasonHoldout
		}, holdoutPropagationTimeout, time.Second,
			"the local evaluation doesn't hold the user out: %s", heldOut[0].ID)
	}
	mismatches := 0
	for _, u := range users {
		evaluation := expected[u.ID]
		details := sdk.StringVariationDetails(ctx, u, featureID, "default")
		if details.VariationID != evaluation.VariationID ||
			details.VariationValue != evaluation.VariationValue ||
			string(details.Reason) != string(evaluation.Reason.Type) ||
			details.FeatureVersion != evaluation.FeatureVersion {
			mismatches++
			t.Errorf(
				"userID: %s\n  server: variation %s (%s), reason %s, version %d\n  local:  variation %s (%s), reason %s, version %d",
				u.ID,
				evaluation.VariationID, evaluation.VariationValue, evaluation.Reason.Type, evaluation.FeatureVersion,
				details.VariationID, details.VariationValue, details.Reason, details.FeatureVersion,
			)
		}
	}
	t.Logf("local: %d of %d users differ from the server", mismatches, len(users))
	// The dropped events are counted before the SDK is closed. Close drains the events before it stops
	// the cache pollers, so a poll that ends in between pushes metrics events that are rejected
	// and counted as dropped, although no evaluation event is lost
	assert.Zero(t, sdk.EventStats().EventsDropped, "an evaluation event was dropped")
	// The evaluation events, which carry the reason and the holdout group ID, are sent here.
	// The server must accept all of them
	require.NoError(t, sdk.Close(ctx))
	stats := sdk.EventStats()
	t.Logf("events: %+v", stats)
	assert.Equal(t, stats.EventsCreated, stats.EventsSent)
	assert.GreaterOrEqual(t, stats.EventsSent, int64(len(users)))

	// The remote evaluation of the SDK reports the reason of the server
	remoteSDK := newSDK(t, ctx)
	check := notHeldOut[0]
	expectedReason := model.EvaluationReason(expected[check.ID].Reason.Type)
	if *holdoutServing {
		check = heldOut[0]
		expectedReason = model.EvaluationReasonHoldout
	}
	details := remoteSDK.StringVariationDetails(ctx, check, featureID, "default")
	assert.Equal(t, expectedReason, details.Reason, "userID: %s", check.ID)
	assert.Equal(t, expected[check.ID].VariationID, details.VariationID, "userID: %s", check.ID)
	require.NoError(t, remoteSDK.Close(ctx))
}
