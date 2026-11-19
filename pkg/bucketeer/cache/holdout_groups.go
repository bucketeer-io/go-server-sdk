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

//go:generate mockgen -source=$GOFILE -package=mock -destination=./mock/$GOFILE
package cache

import (
	"errors"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/v2/proto/feature"
)

const (
	holdoutGroupsKey      = "bucketeer_holdout_groups"
	holdoutGroupsCacheTTL = time.Duration(0)
)

// HoldoutGroupsCache caches the holdout groups that are serving in the environment.
//
// The server always sends the full list, so the list is stored and replaced as a whole.
// An evaluation therefore never sees a list that is half replaced.
type HoldoutGroupsCache interface {
	// Get returns the holdout groups that are serving.
	// The list is empty when no group is serving or when nothing has been cached yet.
	//
	// The returned groups are shared with other callers and must not be modified.
	Get() ([]*ftproto.HoldoutGroup, error)

	// Put replaces the cached holdout groups. An empty list clears them.
	//
	// The cache keeps the groups it is given. The caller must not modify them afterward.
	Put(groups []*ftproto.HoldoutGroup) error
}

type holdoutGroupsCache struct {
	cache Cache
}

func NewHoldoutGroupsCache(c Cache) HoldoutGroupsCache {
	return &holdoutGroupsCache{cache: c}
}

func (c *holdoutGroupsCache) Get() ([]*ftproto.HoldoutGroup, error) {
	value, err := c.cache.Get(holdoutGroupsKey)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	groups, ok := value.([]*ftproto.HoldoutGroup)
	if !ok {
		return nil, ErrInvalidType
	}
	return groups, nil
}

func (c *holdoutGroupsCache) Put(groups []*ftproto.HoldoutGroup) error {
	if len(groups) == 0 {
		c.cache.Delete(holdoutGroupsKey)
		return nil
	}
	return c.cache.Put(holdoutGroupsKey, groups, holdoutGroupsCacheTTL)
}
