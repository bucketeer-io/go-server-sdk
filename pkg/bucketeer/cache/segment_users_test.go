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
			segmentUsers, err := suc.Get(segmentID)
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
			err := suc.Put(p.input)
			assert.Equal(t, p.expected, err)
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
