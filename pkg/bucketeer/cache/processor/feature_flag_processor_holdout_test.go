package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	mockcache "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
	mockevt "github.com/bucketeer-io/go-server-sdk/test/mock/event"
)

func newHoldoutGroupModel(t *testing.T, id string, featureIDs ...string) model.HoldoutGroup {
	t.Helper()
	group := model.HoldoutGroup{
		ID:             id,
		EnvironmentID:  "environment-id",
		AssignmentType: "HASH",
		SamplingSeed:   "sampling-seed",
		BucketStart:    10,
		Percentage:     5,
		Status:         "RUNNING",
		UpdatedAt:      "1763500000",
	}
	for _, featureID := range featureIDs {
		group.Features = append(group.Features, model.HoldoutGroupFeature{
			ID:                 id + ":" + featureID,
			HoldoutGroupID:     id,
			FeatureID:          featureID,
			EnvironmentID:      "environment-id",
			HoldoutVariationID: "variation-id",
			AddedAt:            "1763400000",
		})
	}
	return group
}

// It polls the server several times with the real cache,
// and checks the holdout groups the evaluator would read after each response.
func TestUpdateCacheHoldoutGroups(t *testing.T) {
	t.Parallel()
	internalErr := errors.New("internal error")

	type holdoutGroups map[string][]string // Holdout group ID -> target feature IDs

	steps := []struct {
		desc string
		// The featureFlagsID the SDK must send. It is the ID of the last successful response.
		expectedRequestID string
		response          *model.GetFeatureFlagsResponse
		responseErr       error
		expectedGroups    holdoutGroups
		expectedFlagIDs   []string
		missingFlagIDs    []string
	}{
		{
			desc:              "the first response is a forced update and carries a group",
			expectedRequestID: "",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID: "100:1",
				RequestedAt:    "10",
				ForceUpdate:    true,
				Features:       []model.Feature{{ID: "feature-id-1"}, {ID: "feature-id-2"}},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-1", "feature-id-1"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-1": {"feature-id-1"}},
			expectedFlagIDs: []string{"feature-id-1", "feature-id-2"},
		},
		{
			desc:              "an empty flag diff replaces the list: a target was added",
			expectedRequestID: "100:1",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "100:2",
				RequestedAt:            "10",
				Features:               []model.Feature{},
				ArchivedFeatureFlagIDs: []string{},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-1", "feature-id-1", "feature-id-2"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-1": {"feature-id-1", "feature-id-2"}},
			expectedFlagIDs: []string{"feature-id-1", "feature-id-2"},
		},
		{
			desc:              "an empty flag diff replaces the list: another group is serving",
			expectedRequestID: "100:2",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "100:3",
				RequestedAt:            "10",
				Features:               []model.Feature{},
				ArchivedFeatureFlagIDs: []string{},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-2", "feature-id-2"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-2": {"feature-id-2"}},
			expectedFlagIDs: []string{"feature-id-1", "feature-id-2"},
		},
		{
			desc:              "the response with nothing changed keeps the list it carries",
			expectedRequestID: "100:3",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "100:3",
				RequestedAt:            "10",
				Features:               []model.Feature{},
				ArchivedFeatureFlagIDs: []string{},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-2", "feature-id-2"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-2": {"feature-id-2"}},
			expectedFlagIDs: []string{"feature-id-1", "feature-id-2"},
		},
		{
			desc:              "a flag diff replaces the list too",
			expectedRequestID: "100:3",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "101:4",
				RequestedAt:            "20",
				Features:               []model.Feature{{ID: "feature-id-3"}},
				ArchivedFeatureFlagIDs: []string{"feature-id-1"},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-2", "feature-id-2", "feature-id-3"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-2": {"feature-id-2", "feature-id-3"}},
			expectedFlagIDs: []string{"feature-id-2", "feature-id-3"},
			missingFlagIDs:  []string{"feature-id-1"},
		},
		{
			desc:              "a forced update replaces the list and the flags",
			expectedRequestID: "101:4",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID: "200:5",
				RequestedAt:    "30",
				ForceUpdate:    true,
				Features:       []model.Feature{{ID: "feature-id-4"}},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-3", "feature-id-4"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-3": {"feature-id-4"}},
			expectedFlagIDs: []string{"feature-id-4"},
			missingFlagIDs:  []string{"feature-id-2", "feature-id-3"},
		},
		{
			desc:              "a failed request keeps the list",
			expectedRequestID: "200:5",
			responseErr:       internalErr,
			expectedGroups:    holdoutGroups{"holdout-group-id-3": {"feature-id-4"}},
			expectedFlagIDs:   []string{"feature-id-4"},
		},
		{
			desc:              "an empty list on an empty flag diff clears the list: the group stopped",
			expectedRequestID: "200:5",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "200",
				RequestedAt:            "30",
				Features:               []model.Feature{},
				ArchivedFeatureFlagIDs: []string{},
				HoldoutGroups:          []model.HoldoutGroup{},
			},
			expectedGroups:  holdoutGroups{},
			expectedFlagIDs: []string{"feature-id-4"},
		},
		{
			desc:              "a group is serving again",
			expectedRequestID: "200",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "200:6",
				RequestedAt:            "30",
				Features:               []model.Feature{},
				ArchivedFeatureFlagIDs: []string{},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-4", "feature-id-4"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-4": {"feature-id-4"}},
			expectedFlagIDs: []string{"feature-id-4"},
		},
		{
			desc:              "an absent list on a flag diff clears the list: the server doesn't support holdout groups",
			expectedRequestID: "200:6",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "201",
				RequestedAt:            "40",
				Features:               []model.Feature{{ID: "feature-id-5"}},
				ArchivedFeatureFlagIDs: []string{},
			},
			expectedGroups:  holdoutGroups{},
			expectedFlagIDs: []string{"feature-id-4", "feature-id-5"},
		},
		{
			desc:              "a group is serving again",
			expectedRequestID: "201",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID:         "201:7",
				RequestedAt:            "40",
				Features:               []model.Feature{},
				ArchivedFeatureFlagIDs: []string{},
				HoldoutGroups: []model.HoldoutGroup{
					newHoldoutGroupModel(t, "holdout-group-id-5", "feature-id-5"),
				},
			},
			expectedGroups:  holdoutGroups{"holdout-group-id-5": {"feature-id-5"}},
			expectedFlagIDs: []string{"feature-id-4", "feature-id-5"},
		},
		{
			desc:              "an empty list on a forced update clears the list",
			expectedRequestID: "201:7",
			response: &model.GetFeatureFlagsResponse{
				FeatureFlagsID: "300",
				RequestedAt:    "50",
				ForceUpdate:    true,
				Features:       []model.Feature{{ID: "feature-id-6"}},
				HoldoutGroups:  []model.HoldoutGroup{},
			},
			expectedGroups:  holdoutGroups{},
			expectedFlagIDs: []string{"feature-id-6"},
			missingFlagIDs:  []string{"feature-id-4", "feature-id-5"},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()
	apiClient := mockapi.NewMockClient(controller)
	eventProcessor := mockevt.NewMockProcessor(controller)
	eventProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags).AnyTimes()
	eventProcessor.EXPECT().PushSizeMetricsEvent(gomock.Any(), model.GetFeatureFlags).AnyTimes()
	eventProcessor.EXPECT().PushErrorEvent(gomock.Any(), model.GetFeatureFlags).Times(1)

	inMemoryCache := cache.NewInMemoryCache()
	p := NewFeatureFlagProcessor(&FeatureFlagProcessorConfig{
		Cache:                   inMemoryCache,
		PollingInterval:         time.Minute,
		APIClient:               apiClient,
		PushLatencyMetricsEvent: eventProcessor.PushLatencyMetricsEvent,
		PushSizeMetricsEvent:    eventProcessor.PushSizeMetricsEvent,
		PushErrorEvent:          eventProcessor.PushErrorEvent,
		Loggers: log.NewLoggers(&log.LoggersConfig{
			EnableDebugLog: false,
			ErrorLogger:    log.DiscardErrorLogger,
		}),
		Tag:        "tag",
		SDKVersion: sdkVersion,
		SourceID:   model.SourceIDGoServer,
	}).(*processor)
	// The evaluator reads the same in-memory cache through its own cache instances
	holdoutGroupsCache := cache.NewHoldoutGroupsCache(inMemoryCache)
	featuresCache := cache.NewFeaturesCache(inMemoryCache)

	for _, step := range steps {
		apiClient.EXPECT().GetFeatureFlags(
			gomock.Any(),
			gomock.Cond(func(x any) bool {
				req, ok := x.(*model.GetFeatureFlagsRequest)
				return ok && req.FeatureFlagsID == step.expectedRequestID
			}),
			gomock.Any(),
		).Return(step.response, 1, step.responseErr)

		err := p.updateCache(context.Background())
		require.Equal(t, step.responseErr, err, step.desc)

		groups, err := holdoutGroupsCache.Get()
		require.NoError(t, err, step.desc)
		actualGroups := holdoutGroups{}
		for _, group := range groups {
			// All the fields of the group reach the cache, not only the ID and the targets
			assert.Equal(t, "environment-id", group.EnvironmentId, step.desc)
			assert.Equal(t, ftproto.HoldoutGroup_HASH, group.AssignmentType, step.desc)
			assert.Equal(t, "sampling-seed", group.SamplingSeed, step.desc)
			assert.Equal(t, float64(10), group.BucketStart, step.desc)
			assert.Equal(t, float64(5), group.Percentage, step.desc)
			assert.Equal(t, ftproto.HoldoutGroup_RUNNING, group.Status, step.desc)
			featureIDs := []string{}
			for _, feature := range group.Features {
				assert.Equal(t, "variation-id", feature.HoldoutVariationId, step.desc)
				featureIDs = append(featureIDs, feature.FeatureId)
			}
			actualGroups[group.Id] = featureIDs
		}
		assert.Equal(t, step.expectedGroups, actualGroups, step.desc)

		for _, id := range step.expectedFlagIDs {
			_, err := featuresCache.Get(id)
			assert.NoError(t, err, "%s: %s", step.desc, id)
		}
		for _, id := range step.missingFlagIDs {
			_, err := featuresCache.Get(id)
			assert.ErrorIs(t, err, cache.ErrNotFound, "%s: %s", step.desc, id)
		}
	}
}

// The featureFlagsID covers the holdout groups while a group is serving.
// It must not be stored when the groups it was issued for could not be saved,
// otherwise the server would see an SDK that holds the current state.
func TestUpdateCacheFailedToPutHoldoutGroups(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	defer controller.Finish()
	internalErr := errors.New("internal error")

	patterns := []struct {
		desc        string
		forceUpdate bool
	}{
		{desc: "force update is true", forceUpdate: true},
		{desc: "force update is false", forceUpdate: false},
	}
	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			p := newMockFeatureFlagProcessor(t, controller, "tag", 3*time.Second)
			holdoutGroupsCache := mockcache.NewMockHoldoutGroupsCache(controller)
			p.holdoutGroupsCache = holdoutGroupsCache

			p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("100", nil)
			p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)
			p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
				gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(
				&model.GetFeatureFlagsResponse{
					FeatureFlagsID:         "100:1",
					RequestedAt:            "10",
					ForceUpdate:            pt.forceUpdate,
					Features:               []model.Feature{},
					ArchivedFeatureFlagIDs: []string{},
					HoldoutGroups: []model.HoldoutGroup{
						newHoldoutGroupModel(t, "holdout-group-id", "feature-id"),
					},
				},
				1,
				nil,
			)
			p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
			p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)
			if pt.forceUpdate {
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().DeleteAll().Return(nil)
			}
			holdoutGroupsCache.EXPECT().Put(gomock.Cond(func(x any) bool {
				groups, ok := x.([]*ftproto.HoldoutGroup)
				return ok && len(groups) == 1 && groups[0].Id == "holdout-group-id"
			})).Return(internalErr)
			// The featureFlagsID and the requestedAt are not stored: the mock cache expects no Put
			p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)

			err := p.updateCache(context.Background())
			assert.Equal(t, internalErr, err)
			assert.False(t, p.IsReady())
		})
	}
}
