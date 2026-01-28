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

	maxTimes := 8

	p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsIDKey).Return("", nil).Times(maxTimes)
	p.cache.(*mockcache.MockCache).EXPECT().Get(featureFlagsRequestedAtKey).Return(int64(0), nil).Times(maxTimes)
	p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsIDKey, "", cacheTTL).Return(nil).Times(maxTimes)
	p.cache.(*mockcache.MockCache).EXPECT().Put(featureFlagsRequestedAtKey, int64(0), cacheTTL).Return(nil).Times(maxTimes)

	p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&model.GetFeatureFlagsResponse{},
		1,
		nil,
	).Times(maxTimes)

	p.MockProcessor.EXPECT().PushLatencyMetricsEvent(gomock.Any(), model.GetFeatureFlags).Times(maxTimes)
	p.MockProcessor.EXPECT().PushSizeMetricsEvent(1, model.GetFeatureFlags).Times(maxTimes)

	p.processor.Run()
	time.Sleep(10 * time.Second)
	p.processor.Close()

	// Run it again after closing
	p.Run()
	time.Sleep(10 * time.Second)
	p.processor.Close()
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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

				p.apiClient.(*mockapi.MockClient).EXPECT().GetFeatureFlagsWithDeadline(
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
