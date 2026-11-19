// Copyright 2026 The Bucketeer Authors.
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
	"errors"
	"testing"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	cachemock "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache/mock"
)

func TestHoldoutGroupsCache(t *testing.T) {
	t.Parallel()
	inMemory := NewInMemoryCache()
	c := NewHoldoutGroupsCache(inMemory)
	group1 := &ftproto.HoldoutGroup{Id: "holdout-group-id-1"}
	group2 := &ftproto.HoldoutGroup{Id: "holdout-group-id-2"}

	// Nothing has been cached yet
	groups, err := c.Get()
	require.NoError(t, err)
	assert.Empty(t, groups)

	// Save the list
	require.NoError(t, c.Put([]*ftproto.HoldoutGroup{group1}))
	groups, err = c.Get()
	require.NoError(t, err)
	assert.Equal(t, []*ftproto.HoldoutGroup{group1}, groups)

	// The list is replaced, not merged
	require.NoError(t, c.Put([]*ftproto.HoldoutGroup{group2}))
	groups, err = c.Get()
	require.NoError(t, err)
	assert.Equal(t, []*ftproto.HoldoutGroup{group2}, groups)

	// An empty list clears the cached one
	require.NoError(t, c.Put([]*ftproto.HoldoutGroup{}))
	groups, err = c.Get()
	require.NoError(t, err)
	assert.Empty(t, groups)
	_, err = inMemory.Get(holdoutGroupsKey)
	assert.ErrorIs(t, err, ErrNotFound)

	// A nil list clears the cached one too
	require.NoError(t, c.Put([]*ftproto.HoldoutGroup{group1, group2}))
	require.NoError(t, c.Put(nil))
	groups, err = c.Get()
	require.NoError(t, err)
	assert.Empty(t, groups)
}

// The flags and the segment users are deleted by key prefix on a forced update.
// It must not take the holdout groups with it.
func TestHoldoutGroupsCacheIsNotDeletedWithTheOtherCaches(t *testing.T) {
	t.Parallel()
	inMemory := NewInMemoryCache()
	c := NewHoldoutGroupsCache(inMemory)
	group := &ftproto.HoldoutGroup{Id: "holdout-group-id"}
	require.NoError(t, c.Put([]*ftproto.HoldoutGroup{group}))

	require.NoError(t, NewFeaturesCache(inMemory).DeleteAll())
	require.NoError(t, NewSegmentUsersCache(inMemory).DeleteAll())

	groups, err := c.Get()
	require.NoError(t, err)
	assert.Equal(t, []*ftproto.HoldoutGroup{group}, groups)
}

func TestHoldoutGroupsCacheErrors(t *testing.T) {
	t.Parallel()
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	internalErr := errors.New("internal error")
	groups := []*ftproto.HoldoutGroup{{Id: "holdout-group-id"}}

	patterns := []struct {
		desc        string
		setup       func(*cachemock.MockCache)
		run         func(HoldoutGroupsCache) ([]*ftproto.HoldoutGroup, error)
		expected    []*ftproto.HoldoutGroup
		expectedErr error
	}{
		{
			desc: "get: internal error",
			setup: func(c *cachemock.MockCache) {
				c.EXPECT().Get(holdoutGroupsKey).Return(nil, internalErr)
			},
			run:         func(c HoldoutGroupsCache) ([]*ftproto.HoldoutGroup, error) { return c.Get() },
			expectedErr: internalErr,
		},
		{
			desc: "get: invalid type",
			setup: func(c *cachemock.MockCache) {
				c.EXPECT().Get(holdoutGroupsKey).Return("invalid", nil)
			},
			run:         func(c HoldoutGroupsCache) ([]*ftproto.HoldoutGroup, error) { return c.Get() },
			expectedErr: ErrInvalidType,
		},
		{
			desc: "get: not found is not an error",
			setup: func(c *cachemock.MockCache) {
				c.EXPECT().Get(holdoutGroupsKey).Return(nil, ErrNotFound)
			},
			run: func(c HoldoutGroupsCache) ([]*ftproto.HoldoutGroup, error) { return c.Get() },
		},
		{
			desc: "put: internal error",
			setup: func(c *cachemock.MockCache) {
				c.EXPECT().Put(holdoutGroupsKey, groups, holdoutGroupsCacheTTL).Return(internalErr)
			},
			run:         func(c HoldoutGroupsCache) ([]*ftproto.HoldoutGroup, error) { return nil, c.Put(groups) },
			expectedErr: internalErr,
		},
		{
			desc: "put: an empty list deletes the key",
			setup: func(c *cachemock.MockCache) {
				c.EXPECT().Delete(holdoutGroupsKey)
			},
			run: func(c HoldoutGroupsCache) ([]*ftproto.HoldoutGroup, error) { return nil, c.Put(nil) },
		},
	}
	for _, p := range patterns {
		t.Run(p.desc, func(t *testing.T) {
			mockCache := cachemock.NewMockCache(mockController)
			p.setup(mockCache)
			actual, err := p.run(NewHoldoutGroupsCache(mockCache))
			assert.Equal(t, p.expected, actual)
			assert.Equal(t, p.expectedErr, err)
		})
	}
}
