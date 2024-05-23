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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	cachemock "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
)

func TestGetFeatureFlag(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	features := createFeatures(t, 1)
	id := features[0].Id
	dataFeatures := marshalMessage(t, features[0])
	key := fmt.Sprintf("%s:%s", featureFlagPrefix, id)

	patterns := []struct {
		desc        string
		setup       func(*featuresCache)
		expected    *ftproto.Feature
		expectedErr error
	}{
		{
			desc: "error: not found",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(nil, ErrNotFound)
			},
			expected:    nil,
			expectedErr: ErrNotFound,
		},
		{
			desc: "error: invalid type",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return("test", nil)
			},
			expected:    nil,
			expectedErr: ErrInvalidType,
		},
		{
			desc: "error: failed to unmarshal proto message",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return([]byte{1}, nil)
			},
			expected:    nil,
			expectedErr: ErrFailedToUnmarshalProto,
		},
		{
			desc: "success",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(dataFeatures, nil)
			},
			expected:    features[0],
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			p.setup(fc)
			features, err := fc.GetFeatureFlag(id)
			assert.True(t, proto.Equal(p.expected, features))
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestPutFeaturesFlag(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	features := createFeatures(t, 1)
	dataFeatures := marshalMessage(t, features[0])
	key := fmt.Sprintf("%s:%s", featureFlagPrefix, features[0].Id)

	patterns := []struct {
		desc     string
		setup    func(*featuresCache)
		input    *ftproto.Feature
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
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Put(key, dataFeatures, featureFlagsCacheTTL).Return(nil)
			},
			input:    features[0],
			expected: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			if p.setup != nil {
				p.setup(fc)
			}
			err := fc.PutFeatureFlag(p.input)
			assert.Equal(t, p.expected, err)
		})
	}
}

func TestGetFeatureFlagsID(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	id := "feature-flags-id"
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, featureFlagsIDKey)
	patterns := []struct {
		desc        string
		setup       func(*featuresCache)
		expected    string
		expectedErr error
	}{
		{
			desc: "error: not found",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(nil, ErrNotFound)
			},
			expected:    "",
			expectedErr: ErrNotFound,
		},
		{
			desc: "error: invalid type",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(1, nil)
			},
			expected:    "",
			expectedErr: ErrInvalidType,
		},
		{
			desc: "success",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(id, nil)
			},
			expected:    id,
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			p.setup(fc)
			id, err := fc.GetFeatureFlagsID()
			assert.Equal(t, p.expected, id)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestPutFeatureFlagsID(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	id := "feature-flags-id"
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, featureFlagsIDKey)

	patterns := []struct {
		desc     string
		setup    func(*featuresCache)
		input    string
		expected error
	}{
		{
			desc: "success",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Put(key, id, featureFlagsCacheTTL).Return(nil)
			},
			input:    id,
			expected: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			p.setup(fc)
			err := fc.PutFeatureFlagsID(p.input)
			assert.Equal(t, p.expected, err)
		})
	}
}

func TestGetRequestedAt(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	timestamp := time.Now().Unix()
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, requestedAtKey)

	patterns := []struct {
		desc        string
		setup       func(*featuresCache)
		expected    int64
		expectedErr error
	}{
		{
			desc: "error: not found",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(nil, ErrNotFound)
			},
			expected:    0,
			expectedErr: ErrNotFound,
		},
		{
			desc: "error: invalid type",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(1, nil)
			},
			expected:    0,
			expectedErr: ErrInvalidType,
		},
		{
			desc: "success",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Get(key).Return(timestamp, nil)
			},
			expected:    timestamp,
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			p.setup(fc)
			timestamp, err := fc.GetRequestedAt()
			assert.Equal(t, p.expected, timestamp)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestPutRequestedAt(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	timestamp := time.Now().Unix()
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, requestedAtKey)

	patterns := []struct {
		desc     string
		setup    func(*featuresCache)
		input    int64
		expected error
	}{
		{
			desc: "success",
			setup: func(fc *featuresCache) {
				fc.cache.(*cachemock.MockCache).EXPECT().Put(key, timestamp, featureFlagsCacheTTL).Return(nil)
			},
			input:    timestamp,
			expected: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			p.setup(fc)
			err := fc.PutRequestedAt(p.input)
			assert.Equal(t, p.expected, err)
		})
	}
}

func createFeatures(t *testing.T, size int) []*ftproto.Feature {
	t.Helper()
	f := []*ftproto.Feature{}
	for i := 0; i < size; i++ {
		feature := &ftproto.Feature{
			Id:   fmt.Sprintf("feature-id-%d", i),
			Name: fmt.Sprintf("feature-name-%d", i),
		}
		f = append(f, feature)
	}
	return f
}

func marshalMessage(t *testing.T, pb proto.Message) interface{} {
	t.Helper()
	buffer, err := proto.Marshal(pb)
	require.NoError(t, err)
	return buffer
}

func newFeaturesCache(t *testing.T, mockController *gomock.Controller) *featuresCache {
	t.Helper()
	return &featuresCache{
		cache: cachemock.NewMockCache(mockController),
	}
}
