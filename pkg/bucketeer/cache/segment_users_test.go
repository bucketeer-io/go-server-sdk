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

package cache

import (
	"fmt"
	"testing"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	cachemock "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
)

func TestGetSegmentUsers(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	segmentID := "segment-id"
	segmentUsers := createSegmentUsers(t, segmentID, 3)
	dataSegmentUsers := marshalMessage(t, segmentUsers)
	key := fmt.Sprintf("%s:%s", segmentUsersFlagsPrefix, segmentID)

	patterns := []struct {
		desc        string
		setup       func(*segmentUsersCache)
		expected    *ftproto.SegmentUsers
		expectedErr error
	}{
		{
			desc: "error: not found",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(nil, ErrNotFound)
			},
			expected:    nil,
			expectedErr: ErrNotFound,
		},
		{
			desc: "error: invalid type",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return("test", nil)
			},
			expected:    nil,
			expectedErr: ErrInvalidType,
		},
		{
			desc: "error: failed to unmarshal proto message",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return([]byte{1}, nil)
			},
			expected:    nil,
			expectedErr: ErrFailedToUnmarshalProto,
		},
		{
			desc: "success",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(dataSegmentUsers, nil)
			},
			expected:    segmentUsers,
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			suc := newMockSegmentUsersCache(t, mockController)
			p.setup(suc)
			segmentUsers, err := suc.GetSegmentUsers(segmentID)
			assert.True(t, proto.Equal(p.expected, segmentUsers))
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestPutSegmentUsers(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	segmentID := "segment-id"
	segmentUsers := createSegmentUsers(t, segmentID, 3)
	dataSegmentUsers := marshalMessage(t, segmentUsers)
	key := fmt.Sprintf("%s:%s", segmentUsersFlagsPrefix, segmentID)

	patterns := []struct {
		desc     string
		setup    func(*segmentUsersCache)
		input    *ftproto.SegmentUsers
		expected error
	}{
		{
			desc:     "error: proto message is nil",
			setup:    nil,
			input:    nil,
			expected: ErrProtoMessageNil,
		},
		{
			desc: "success",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Put(key, dataSegmentUsers, segmentUsersFlagsTTL).Return(nil)
			},
			input:    segmentUsers,
			expected: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			suc := newMockSegmentUsersCache(t, mockController)
			if p.setup != nil {
				p.setup(suc)
			}
			err := suc.PutSegmentUsers(p.input)
			assert.Equal(t, p.expected, err)
		})
	}
}

func TestGetSegmentUsersRequestedAt(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	timestamp := time.Now().Unix()

	patterns := []struct {
		desc        string
		setup       func(*segmentUsersCache)
		expected    int64
		expectedErr error
	}{
		{
			desc: "error: not found",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(segmentUserRequestedAtKey).Return(nil, ErrNotFound)
			},
			expected:    0,
			expectedErr: ErrNotFound,
		},
		{
			desc: "error: invalid type",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(segmentUserRequestedAtKey).Return(1, nil)
			},
			expected:    0,
			expectedErr: ErrInvalidType,
		},
		{
			desc: "success",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Get(segmentUserRequestedAtKey).Return(timestamp, nil)
			},
			expected:    timestamp,
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			suc := newMockSegmentUsersCache(t, mockController)
			p.setup(suc)
			timestamp, err := suc.GetRequestedAt()
			assert.Equal(t, p.expected, timestamp)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestPutSegmentUsersRequestedAt(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	timestamp := time.Now().Unix()

	patterns := []struct {
		desc     string
		setup    func(*segmentUsersCache)
		input    int64
		expected error
	}{
		{
			desc: "success",
			setup: func(suc *segmentUsersCache) {
				suc.cache.(*cachemock.MockCache).EXPECT().Put(segmentUserRequestedAtKey, timestamp, segmentUsersFlagsTTL).Return(nil)
			},
			input:    timestamp,
			expected: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			suc := newMockSegmentUsersCache(t, mockController)
			p.setup(suc)
			err := suc.PutRequestedAt(p.input)
			assert.Equal(t, p.expected, err)
		})
	}
}

func TestGetSegmentIDs(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	segmentUsers1 := createSegmentUsers(t, "segment-id-1", 3)
	segmentUsers2 := createSegmentUsers(t, "segment-id-2", 3)
	segmentUsers3 := createSegmentUsers(t, "segment-id-3", 3)

	patterns := []struct {
		desc        string
		setup       func(*segmentUsersCache)
		expected    []string
		expectedErr error
	}{
		{
			desc: "success: key not string",
			setup: func(suc *segmentUsersCache) {
				suc.cache.Put(1, "random-value", 0)
			},
			expected:    make([]string, 0),
			expectedErr: nil,
		},
		{
			desc: "success: random data",
			setup: func(suc *segmentUsersCache) {
				suc.cache.Put("random-id", "random-value", 0)
			},
			expected:    make([]string, 0),
			expectedErr: nil,
		},
		{
			desc: "success",
			setup: func(suc *segmentUsersCache) {
				suc.PutSegmentUsers(segmentUsers1)
				suc.PutSegmentUsers(segmentUsers2)
				suc.PutSegmentUsers(segmentUsers3)
			},
			expected:    []string{"segment-id-1", "segment-id-2", "segment-id-3"},
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			suc := newSegmentUsersCache(t)
			p.setup(suc)
			ids, err := suc.GetSegmentIDs()
			assert.True(t, len(p.expected) == len(ids))
			assert.Equal(t, p.expectedErr, err)
			// Because the cache uses a map, the key list order might change during the tests
			for _, id := range p.expected {
				assert.Contains(t, ids, id)
			}
		})
	}
}

func createSegmentUsers(t *testing.T, segmentID string, size int) *ftproto.SegmentUsers {
	t.Helper()
	users := make([]*ftproto.SegmentUser, 0, size)
	for i := 0; i < size; i++ {
		user := &ftproto.SegmentUser{
			Id:        fmt.Sprintf("id-%d", i),
			SegmentId: segmentID,
			UserId:    fmt.Sprintf("user-id-%d", i),
		}
		users = append(users, user)
	}
	return &ftproto.SegmentUsers{
		SegmentId: segmentID,
		Users:     users,
		UpdatedAt: 1,
	}
}

func newMockSegmentUsersCache(t *testing.T, mockController *gomock.Controller) *segmentUsersCache {
	t.Helper()
	return &segmentUsersCache{
		cache: cachemock.NewMockCache(mockController),
	}
}

func newSegmentUsersCache(t *testing.T) *segmentUsersCache {
	t.Helper()
	return &segmentUsersCache{
		cache: NewInMemoryCache(),
	}
}
