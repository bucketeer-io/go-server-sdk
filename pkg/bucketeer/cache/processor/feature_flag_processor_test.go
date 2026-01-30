// Copyright 2024 The Bucketeer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	mockcache "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
	mockevt "github.com/bucketeer-io/go-server-sdk/test/mock/event"
)

const (
	sdkVersion = "1.5.5"
)

func TestPollingInterval(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	p := newMockFeatureFlagProcessor(
		t,
		mockController,
		"tag",
		3*time.Second,
	)

	// Test runs for 10 seconds with 3s polling interval
	// Expected polls per run: 3-5 (timing dependent)
	// - Minimum: T=0, T=3, T=6 = 3 polls
	// - Maximum: T=0, T=3, T=6, T=9, T=12 = 5 polls
	// Two runs total, so 6-10 polls total
	minPolls := 6
	maxPolls := 10

	// Each updateCache() call makes:
	// 1. checkAndHealStaleCache() -> Get(requestedAt) IF ready=true
	// 2. getFeatureFlagsID() -> Get(featureFlagsIDKey)
	// 3. getFeatureFlagsRequestedAt() -> Get(requestedAt)
	//
	// First poll of each run: ready=false, so only 1 Get(requestedAt)
	// Subsequent polls: ready=true, so 2 Get(requestedAt) each
	//
	// Min: 2 first polls (2*1) + 4 subsequent polls (4*2) = 10 calls
	// Max: 2 first polls (2*1) + 8 subsequent polls (8*2) = 18 calls

	// Return fresh timestamp on each call to prevent healing from triggering
	p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).DoAndReturn(func(key string) (interface{}, error) {
		return time.Now().Unix(), nil
	}).MinTimes(10).MaxTimes(18)
	p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("", nil).MinTimes(minPolls).MaxTimes(maxPolls)
	p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "", cacheTTL).Return(nil).MinTimes(minPolls).MaxTimes(maxPolls)
	p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(0), cacheTTL).Return(nil).MinTimes(minPolls).MaxTimes(maxPolls)

	p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&model.GetFeatureFlagsResponse{},
		1,
		nil,
	).MinTimes(minPolls).MaxTimes(maxPolls)

	p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags).MinTimes(minPolls).MaxTimes(maxPolls)
	p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags).MinTimes(minPolls).MaxTimes(maxPolls)

	p.processor.Run()
	time.Sleep(10 * time.Second)
	p.processor.Close()
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to fully stop

	// Run it again after closing
	p.Run()
	time.Sleep(10 * time.Second)
	p.processor.Close()
	time.Sleep(100 * time.Millisecond) // Wait for goroutine to fully stop
}

func TestUpdateCache(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	tag := "tag"
	archivedFlagIDs := []string{
		"feature-flags-id-3",
		"feature-flags-id-4",
	}
	singleFeature := model.Feature{ID: "feature-flag-id-2"}
	internalErr := errors.New("internal error")

	patterns := []struct {
		desc         string
		setup        func(*testFeatureFlagProcessor)
		tag          string
		pollInterval time.Duration
		expected     error
	}{
		{
			desc: "err: failed while getting featureFlagsID",
			setup: func(p *testFeatureFlagProcessor) {
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("", internalErr)
				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "err: failed while getting requestedAt",
			setup: func(p *testFeatureFlagProcessor) {
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return("", internalErr)
				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "err: failed while requesting cache from the server",
			setup: func(p *testFeatureFlagProcessor) {
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "" && req.SDKVersion == sdkVersion
					}),
					gomock.Any(),
				).Return(
					nil,
					0,
					internalErr,
				)

				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "err: failed while putting featureFlagsID, and the forceUpdate is true",
			setup: func(p *testFeatureFlagProcessor) {
				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "feature-flags-id-2", cacheTTL).Return(internalErr)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{
						FeatureFlagsID:         "feature-flags-id-2",
						RequestedAt:            "20",
						Features:               []model.Feature{singleFeature},
						ForceUpdate:            true,
						ArchivedFeatureFlagIDs: make([]string, 0),
					},
					1,
					nil,
				)
				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)

				// Call in the feature flag cache
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().DeleteAll().Return(nil)
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(
					gomock.Cond(func(x any) bool {
						f := x.(*ftproto.Feature)
						return f.Id == "feature-flag-id-2"
					})).Return(nil)

				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "err: failed while putting requestedAt, and the forceUpdate is true",
			setup: func(p *testFeatureFlagProcessor) {
				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{
						FeatureFlagsID:         "feature-flags-id-2",
						RequestedAt:            "20",
						Features:               []model.Feature{singleFeature},
						ForceUpdate:            true,
						ArchivedFeatureFlagIDs: make([]string, 0),
					},
					1,
					nil,
				)
				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)

				// Call in the feature flag cache
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().DeleteAll().Return(nil)
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(
					gomock.Cond(func(x any) bool {
						f := x.(*ftproto.Feature)
						return f.Id == "feature-flag-id-2"
					})).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "feature-flags-id-2", cacheTTL).Return(nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(20), cacheTTL).Return(internalErr)

				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "err: failed while putting featureFlagsID, and force update is false",
			setup: func(p *testFeatureFlagProcessor) {
				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{
						FeatureFlagsID:         "feature-flags-id-2",
						RequestedAt:            "20",
						Features:               []model.Feature{singleFeature},
						ForceUpdate:            false,
						ArchivedFeatureFlagIDs: archivedFlagIDs,
					},
					1,
					nil,
				)
				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)

				// Call in the feature flag cache
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(gomock.Cond(func(x any) bool {
					f := x.(*ftproto.Feature)
					return f.Id == "feature-flag-id-2"
				})).Return(nil)
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete(archivedFlagIDs[0])
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete(archivedFlagIDs[1])

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "feature-flags-id-2", cacheTTL).Return(internalErr)

				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "err: failed while putting requestedAt, and force update is false",
			setup: func(p *testFeatureFlagProcessor) {
				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{
						FeatureFlagsID:         "feature-flags-id-2",
						RequestedAt:            "20",
						Features:               []model.Feature{singleFeature},
						ForceUpdate:            false,
						ArchivedFeatureFlagIDs: archivedFlagIDs,
					},
					1,
					nil,
				)
				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)

				// Call in the feature flag cache
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(gomock.Cond(func(x any) bool {
					f := x.(*ftproto.Feature)
					return f.Id == "feature-flag-id-2"
				})).Return(nil)
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete(archivedFlagIDs[0])
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete(archivedFlagIDs[1])

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "feature-flags-id-2", cacheTTL).Return(nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(20), cacheTTL).Return(internalErr)

				p.MockProcessor.EXPECT().PushErrorEvent(p.newInternalError(internalErr), model.GetFeatureFlags)
			},
			tag:      tag,
			expected: internalErr,
		},
		{
			desc: "success: featureFlagsID not found",
			setup: func(p *testFeatureFlagProcessor) {
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("", cache.ErrNotFound)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "", cacheTTL).Return(nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(0), cacheTTL).Return(nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == ""
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{},
					1,
					nil,
				)

				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)
			},
			tag:      tag,
			expected: nil,
		},
		{
			desc: "success: requestedAt not found",
			setup: func(p *testFeatureFlagProcessor) {
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), cache.ErrNotFound)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "", cacheTTL).Return(nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(0), cacheTTL).Return(nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{},
					1,
					nil,
				)

				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)
			},
			tag:      tag,
			expected: nil,
		},
		{
			desc: "success: force update is true",
			setup: func(p *testFeatureFlagProcessor) {
				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)

				// Call in the feature flag cache
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().DeleteAll().Return(nil)
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(gomock.Cond(func(x any) bool {
					f := x.(*ftproto.Feature)
					return f.Id == "feature-flag-id-2"
				})).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "feature-flags-id-2", cacheTTL).Return(nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(20), cacheTTL).Return(nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{
						FeatureFlagsID:         "feature-flags-id-2",
						RequestedAt:            "20",
						Features:               []model.Feature{singleFeature},
						ForceUpdate:            true,
						ArchivedFeatureFlagIDs: make([]string, 0),
					},
					1,
					nil,
				)

				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)
			},
			tag:      tag,
			expected: nil,
		},
		{
			desc: "success: force update is false",
			setup: func(p *testFeatureFlagProcessor) {
				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("feature-flags-id-1", nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(10), nil)

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
					gomock.Any(),
					gomock.Cond(func(x any) bool {
						req, ok := x.(*model.GetFeatureFlagsRequest)
						return ok && req.Tag == tag && req.FeatureFlagsID == "feature-flags-id-1"
					}),
					gomock.Any(),
				).Return(
					&model.GetFeatureFlagsResponse{
						FeatureFlagsID:         "feature-flags-id-2",
						RequestedAt:            "20",
						Features:               []model.Feature{singleFeature},
						ForceUpdate:            false,
						ArchivedFeatureFlagIDs: archivedFlagIDs,
					},
					1,
					nil,
				)

				p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
				p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)

				// Call in the feature flag cache
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(gomock.Cond(func(x any) bool {
					f := x.(*ftproto.Feature)
					return f.Id == "feature-flag-id-2"
				})).Return(nil)
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete(archivedFlagIDs[0])
				p.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete(archivedFlagIDs[1])

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "feature-flags-id-2", cacheTTL).Return(nil)
				p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(20), cacheTTL).Return(nil)
			},
			tag:      tag,
			expected: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			processor := newMockFeatureFlagProcessor(
				t,
				mockController,
				p.tag,
				3*time.Second,
			)
			p.setup(processor)
			err := processor.updateCache(context.Background())
			assert.Equal(t, p.expected, err)
		})
	}
}

type testFeatureFlagProcessor struct {
	*processor
	*mockevt.MockProcessor
}

func newMockFeatureFlagProcessor(
	t *testing.T,
	controller *gomock.Controller,
	tag string,
	pollingInterval time.Duration,
) *testFeatureFlagProcessor {
	t.Helper()
	cacheInMemory := mockcache.NewMockCache(controller)
	featureFlagsCache := mockcache.NewMockFeaturesCache(controller)
	loggerConf := &log.LoggersConfig{
		EnableDebugLog: true,
		ErrorLogger:    log.DefaultErrorLogger,
	}
	mockEventProcessor := mockevt.NewMockProcessor(controller)
	return &testFeatureFlagProcessor{
		processor: &processor{
			apiClient:               mockapi.NewMockClient(controller),
			pushLatencyMetricsEvent: mockEventProcessor.PushLatencyMetricsEvent,
			pushSizeMetricsEvent:    mockEventProcessor.PushSizeMetricsEvent,
			pushErrorEvent:          mockEventProcessor.PushErrorEvent,
			cache:                   cacheInMemory,
			featureFlagsCache:       featureFlagsCache,
			tag:                     tag,
			sdkVersion:              sdkVersion,
			sourceID:                model.SourceIDGoServer,
			closeCh:                 make(chan struct{}),
			pollingInterval:         pollingInterval,
			loggers:                 log.NewLoggers(loggerConf),
		},
		MockProcessor: mockEventProcessor,
	}
}

func TestCheckAndHealStaleCache(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	patterns := []struct {
		desc                           string
		setup                          func(*testFeatureFlagProcessor)
		expectedHealingInProgress      bool
		expectedReady                  bool
		expectMetadataDeleted          bool
		expectHealingInProgressChanged bool
	}{
		{
			desc: "no action: ready is false (not initialized)",
			setup: func(p *testFeatureFlagProcessor) {
				// Don't set ready to true
				p.ready.Store(false)
			},
			expectedReady:             false,
			expectedHealingInProgress: false,
			expectMetadataDeleted:     false,
		},
		{
			desc: "no action: cache is fresh (within threshold)",
			setup: func(p *testFeatureFlagProcessor) {
				p.ready.Store(true)
				// Recent timestamp (10 seconds old, threshold is 120s for 60s interval)
				recentTime := time.Now().Unix() - 10
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(recentTime, nil)
			},
			expectedReady:             true,
			expectedHealingInProgress: false,
			expectMetadataDeleted:     false,
		},
		{
			desc: "triggers healing: cache is stale (beyond threshold)",
			setup: func(p *testFeatureFlagProcessor) {
				p.ready.Store(true)
				// Old timestamp (150 seconds old, threshold is 120s for 60s interval)
				staleTime := time.Now().Unix() - 150
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(staleTime, nil)

				// Expect metadata to be deleted
				p.cache.(*mockcache.MockCache).EXPECT().Delete(featureFlagsIDKey)
				p.cache.(*mockcache.MockCache).EXPECT().Delete(featureFlagsRequestedAtKey)
			},
			expectedReady:             true,
			expectedHealingInProgress: true,
			expectMetadataDeleted:     true,
		},
		{
			desc: "no action: healing already in progress",
			setup: func(p *testFeatureFlagProcessor) {
				p.ready.Store(true)
				p.healingInProgress.Store(true)
				// Metadata missing but healing in progress (expected state)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), cache.ErrNotFound)
			},
			expectedReady:             true,
			expectedHealingInProgress: true,
			expectMetadataDeleted:     false,
		},
		{
			desc: "triggers healing: unexpected missing metadata",
			setup: func(p *testFeatureFlagProcessor) {
				p.ready.Store(true)
				p.healingInProgress.Store(false)
				// Metadata missing but healing NOT in progress (unexpected state)
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), cache.ErrNotFound)

				// Now also deletes metadata to ensure full refresh (consistency fix)
				p.cache.(*mockcache.MockCache).EXPECT().Delete(featureFlagsIDKey)
				p.cache.(*mockcache.MockCache).EXPECT().Delete(featureFlagsRequestedAtKey)
			},
			expectedReady:             true,
			expectedHealingInProgress: true,
			expectMetadataDeleted:     true,
		},
		{
			desc: "triggers healing: corrupted metadata (wrong type)",
			setup: func(p *testFeatureFlagProcessor) {
				p.ready.Store(true)
				p.healingInProgress.Store(false)
				// Metadata exists but corrupted to wrong type
				p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return("corrupted-string", nil)

				// Now also deletes metadata to ensure full refresh (consistency fix)
				p.cache.(*mockcache.MockCache).EXPECT().Delete(featureFlagsIDKey)
				p.cache.(*mockcache.MockCache).EXPECT().Delete(featureFlagsRequestedAtKey)
			},
			expectedReady:             true,
			expectedHealingInProgress: true,
			expectMetadataDeleted:     true,
		},
	}

	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			processor := newMockFeatureFlagProcessor(
				t,
				mockController,
				"tag",
				60*time.Second, // 60s polling interval
			)

			pt.setup(processor)

			err := processor.checkAndHealStaleCache()
			assert.NoError(t, err)
			assert.Equal(t, pt.expectedReady, processor.ready.Load())
			assert.Equal(t, pt.expectedHealingInProgress, processor.healingInProgress.Load())
		})
	}
}

func TestHealingFlagClearedOnSuccess(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	patterns := []struct {
		desc        string
		forceUpdate bool
	}{
		{
			desc:        "healing flag cleared after successful force update",
			forceUpdate: true,
		},
		{
			desc:        "healing flag cleared after successful incremental update",
			forceUpdate: false,
		},
	}

	for _, pt := range patterns {
		t.Run(pt.desc, func(t *testing.T) {
			processor := newMockFeatureFlagProcessor(
				t,
				mockController,
				"tag",
				60*time.Second,
			)

			// Set up initial state: ready and healing in progress
			processor.ready.Store(true)
			processor.healingInProgress.Store(true)

			// Mock checkAndHealStaleCache call at the beginning of updateCache
			// With healing in progress, it expects to find missing metadata
			processor.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), cache.ErrNotFound)

			// Mock the rest of updateCache
			processor.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("old-id", nil)
			processor.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(100), nil)

			processor.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).Return(
				&model.GetFeatureFlagsResponse{
					FeatureFlagsID:         "new-id",
					RequestedAt:            "123",
					Features:               []model.Feature{{ID: "feature-1"}},
					ForceUpdate:            pt.forceUpdate,
					ArchivedFeatureFlagIDs: []string{"archived-flag-1"},
				},
				1,
				nil,
			)

			processor.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags)
			processor.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags)

			if pt.forceUpdate {
				// Force update path
				processor.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().DeleteAll().Return(nil)
				processor.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(gomock.Any()).Return(nil)
			} else {
				// Incremental update path
				processor.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Put(gomock.Any()).Return(nil)
				processor.featureFlagsCache.(*mockcache.MockFeaturesCache).EXPECT().Delete("archived-flag-1")
			}

			processor.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "new-id", cacheTTL).Return(nil)
			processor.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(123), cacheTTL).Return(nil)

			// Execute updateCache
			err := processor.updateCache(context.Background())

			// Assert
			assert.NoError(t, err)
			assert.True(t, processor.ready.Load(), "ready should be true")
			assert.False(t, processor.healingInProgress.Load(), "healingInProgress should be cleared after success")
		})
	}
}

func TestHealingFlagNotClearedOnFailure(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	processor := newMockFeatureFlagProcessor(
		t,
		mockController,
		"tag",
		60*time.Second,
	)

	// Set up initial state: healing in progress
	processor.ready.Store(true)
	processor.healingInProgress.Store(true)

	// Mock checkAndHealStaleCache call - should skip due to healing in progress
	processor.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), cache.ErrNotFound)

	// Mock the rest of updateCache to fail
	processor.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("old-id", nil)
	processor.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(100), nil)

	internalErr := errors.New("network error")
	processor.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlags(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(nil, 0, internalErr)

	processor.MockProcessor.EXPECT().PushErrorEvent(gomock.Any(), model.GetFeatureFlags)

	// Execute updateCache
	err := processor.updateCache(context.Background())

	// Assert
	assert.Error(t, err)
	assert.True(t, processor.ready.Load(), "ready should still be true")
	assert.True(t, processor.healingInProgress.Load(), "healingInProgress should NOT be cleared on failure")
}
