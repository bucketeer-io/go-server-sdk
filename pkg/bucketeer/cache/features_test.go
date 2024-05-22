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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"

	cachemock "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
)

const (
	tag = "bucketeer-tag"
)

func TestGetFeatures(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	features := createFeatures(t)
	dataFeatures := marshalMessage(t, features)
	key := fmt.Sprintf("%s:%s", featureFlagsKind, tag)

	patterns := []struct {
		desc        string
		setup       func(*featuresCache)
		expected    *ftproto.Features
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
			expected:    features,
			expectedErr: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			p.setup(fc)
			features, err := fc.Get(tag)
			assert.True(t, proto.Equal(p.expected, features))
			assert.Equal(t, p.expectedErr, err)
		})
	}
}

func TestPutFeatures(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	features := createFeatures(t)
	dataFeatures := marshalMessage(t, features)
	key := fmt.Sprintf("%s:%s", featureFlagsKind, tag)

	patterns := []struct {
		desc     string
		setup    func(*featuresCache)
		input    *ftproto.Features
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
				fc.cache.(*cachemock.MockCache).EXPECT().Put(key, dataFeatures, featureFlagsTTL).Return(nil)
			},
			input:    features,
			expected: nil,
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			fc := newFeaturesCache(t, mockController)
			if p.setup != nil {
				p.setup(fc)
			}
			err := fc.Put(tag, p.input)
			assert.Equal(t, p.expected, err)
		})
	}
}

func createFeatures(t *testing.T) *ftproto.Features {
	t.Helper()
	f := []*ftproto.Feature{}
	for i := 0; i < 5; i++ {
		feature := &ftproto.Feature{
			Id:   fmt.Sprintf("feature-id-%d", i),
			Name: fmt.Sprintf("feature-name-%d", i),
		}
		f = append(f, feature)
	}
	return &ftproto.Features{
		Id:       "feature-flags-id",
		Features: f,
	}
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
