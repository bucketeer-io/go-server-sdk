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
	"errors"
	"testing"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	gwproto "github.com/bucketeer-io/bucketeer/proto/gateway"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	mockcache "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/log"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	mockapi "github.com/bucketeer-io/go-server-sdk/test/mock/api"
)

func TestSegmentUsersPollingInterval(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	p := newMockSegmentUserProcessor(
		t,
		mockController,
		3*time.Second,
	)

	maxTimes := 8

	p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().Return(make([]string, 0), nil).Times(maxTimes)
	p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(0), nil).Times(maxTimes)
	p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(0), segmentUserCacheTTL).Return(nil).Times(maxTimes)

	p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(gomock.Any()).Return(
		&gwproto.GetSegmentUsersResponse{},
		1,
		nil,
	).Times(maxTimes)

	p.Run()
	time.Sleep(10 * time.Second)
	p.Close()

	// Run it again after closing
	p.Run()
	time.Sleep(10 * time.Second)
	p.Close()
}

func TestSegmentUsersUpdateCache(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	deletedSegmentIDs := []string{
		"segment-id-3",
		"segment-id-4",
	}
	singleSegmentUser := &ftproto.SegmentUsers{
		SegmentId: "segment-id",
		Users: []*ftproto.SegmentUser{
			{
				Id: "user-id",
			},
		},
		UpdatedAt: 20,
	}
	internalErr := errors.New("internal error")

	patterns := []struct {
		desc         string
		setup        func(*segmentUserProcessor)
		pollInterval time.Duration
		expected     error
	}{
		{
			desc: "err: failed while getting segment IDs",
			setup: func(p *segmentUserProcessor) {
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().Return(nil, internalErr)
			},

			expected: errors.New("internal error"),
		},
		{
			desc: "err: failed while getting requestedAt",
			setup: func(p *segmentUserProcessor) {
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().Return(make([]string, 0), nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(0), internalErr)
			},
			expected: errors.New("internal error"),
		},
		{
			desc: "err: failed while requesting cache from the server",
			setup: func(p *segmentUserProcessor) {
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().Return(make([]string, 0), nil)
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(10), nil)

				req := model.NewGetSegmentUsersRequest(make([]string, 0), int64(10))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					nil,
					0,
					internalErr,
				)
			},
			expected: internalErr,
		},
		{
			desc: "err: failed while putting requestedAt, and the forceUpdate is true",
			setup: func(p *segmentUserProcessor) {
				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().
					Return([]string{"segment-id"}, nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(10), nil)

				req := model.NewGetSegmentUsersRequest([]string{"segment-id"}, int64(10))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					&gwproto.GetSegmentUsersResponse{
						SegmentUsers:      []*ftproto.SegmentUsers{singleSegmentUser},
						RequestedAt:       int64(20),
						ForceUpdate:       true,
						DeletedSegmentIds: make([]string, 0),
					},
					1,
					nil,
				)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().DeleteAll().Return(nil)
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Put(singleSegmentUser).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(20), segmentUserCacheTTL).
					Return(internalErr)
			},
			expected: internalErr,
		},
		{
			desc: "err: failed while putting requestedAt, and force update is false",
			setup: func(p *segmentUserProcessor) {
				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().
					Return([]string{"segment-id"}, nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(10), nil)

				req := model.NewGetSegmentUsersRequest([]string{"segment-id"}, int64(10))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					&gwproto.GetSegmentUsersResponse{
						SegmentUsers:      []*ftproto.SegmentUsers{singleSegmentUser},
						RequestedAt:       int64(20),
						ForceUpdate:       false,
						DeletedSegmentIds: make([]string, 0),
					},
					1,
					nil,
				)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Put(singleSegmentUser).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(20), segmentUserCacheTTL).
					Return(internalErr)
			},
			expected: internalErr,
		},
		{
			desc: "success: get segment IDs not found",
			setup: func(p *segmentUserProcessor) {
				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().Return(make([]string, 0), cache.ErrNotFound)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(10), nil)

				req := model.NewGetSegmentUsersRequest(make([]string, 0), int64(10))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					&gwproto.GetSegmentUsersResponse{
						SegmentUsers:      []*ftproto.SegmentUsers{singleSegmentUser},
						RequestedAt:       int64(20),
						ForceUpdate:       false,
						DeletedSegmentIds: make([]string, 0),
					},
					1,
					nil,
				)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Put(singleSegmentUser).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(20), segmentUserCacheTTL).
					Return(nil)
			},
			expected: nil,
		},
		{
			desc: "success: requestedAt not found",
			setup: func(p *segmentUserProcessor) {
				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().Return(make([]string, 0), nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(0), cache.ErrNotFound)

				req := model.NewGetSegmentUsersRequest(make([]string, 0), int64(0))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					&gwproto.GetSegmentUsersResponse{
						SegmentUsers:      []*ftproto.SegmentUsers{singleSegmentUser},
						RequestedAt:       int64(20),
						ForceUpdate:       false,
						DeletedSegmentIds: make([]string, 0),
					},
					1,
					nil,
				)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Put(singleSegmentUser).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(20), segmentUserCacheTTL).
					Return(nil)
			},
			expected: nil,
		},
		{
			desc: "success: force update is true",
			setup: func(p *segmentUserProcessor) {
				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().
					Return([]string{"segment-id"}, nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(10), nil)

				req := model.NewGetSegmentUsersRequest([]string{"segment-id"}, int64(10))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					&gwproto.GetSegmentUsersResponse{
						SegmentUsers:      []*ftproto.SegmentUsers{singleSegmentUser},
						RequestedAt:       int64(20),
						ForceUpdate:       true,
						DeletedSegmentIds: make([]string, 0),
					},
					1,
					nil,
				)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().DeleteAll().Return(nil)
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Put(singleSegmentUser).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(20), segmentUserCacheTTL).
					Return(nil)
			},
			expected: nil,
		},
		{
			desc: "success: force update is false",
			setup: func(p *segmentUserProcessor) {
				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().GetSegmentIDs().
					Return([]string{"segment-id"}, nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Get(segmentUsersRequestedAtKey).Return(int64(10), nil)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Delete(deletedSegmentIDs[0])
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Delete(deletedSegmentIDs[1])

				req := model.NewGetSegmentUsersRequest([]string{"segment-id"}, int64(10))
				p.apiClient.(*mockapi.MockClient).EXPECT().GetSegmentUsers(req).Return(
					&gwproto.GetSegmentUsersResponse{
						SegmentUsers:      []*ftproto.SegmentUsers{singleSegmentUser},
						RequestedAt:       int64(20),
						ForceUpdate:       false,
						DeletedSegmentIds: deletedSegmentIDs,
					},
					1,
					nil,
				)

				// Call in the segment users cache
				p.segmentUsersCache.(*mockcache.MockSegmentUsersCache).EXPECT().Put(singleSegmentUser).Return(nil)

				// Call in the processor cache
				p.cache.(*mockcache.MockCache).EXPECT().Put(segmentUsersRequestedAtKey, int64(20), segmentUserCacheTTL).
					Return(nil)
			},
			expected: nil,
		},
	}

	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			processor := newMockSegmentUserProcessor(
				t,
				mockController,
				3*time.Second,
			)
			p.setup(processor)
			err := processor.updateCache()
			assert.Equal(t, p.expected, err)
		})
	}
}

func newMockSegmentUserProcessor(
	t *testing.T,
	controller *gomock.Controller,
	pollingInterval time.Duration,
) *segmentUserProcessor {
	t.Helper()
	cacheInMemory := mockcache.NewMockCache(controller)
	segmentUsersCache := mockcache.NewMockSegmentUsersCache(controller)
	loggerConf := &log.LoggersConfig{
		EnableDebugLog: true,
		ErrorLogger:    log.DefaultErrorLogger,
	}
	return &segmentUserProcessor{
		apiClient:         mockapi.NewMockClient(controller),
		cache:             cacheInMemory,
		segmentUsersCache: segmentUsersCache,
		closeCh:           make(chan struct{}),
		pollingInterval:   pollingInterval,
		loggers:           log.NewLoggers(loggerConf),
	}
}