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

//go:generate mockgen -source=$GOFILE -package=mock -destination=./mock/$GOFILE
package cache

import (
	"fmt"
	"strings"
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	"google.golang.org/protobuf/proto"
)

const (
	segmentUsersPrefix = "bucketeer_segment_users"
	segmentUsersTTL    = time.Duration(0)
)

type SegmentUsersCache interface {
	// Get the segment users by segment ID
	Get(segmentID string) (*ftproto.SegmentUsers, error)

	// Get all the segment IDs
	GetSegmentIDs() ([]string, error)

	// Save the segment users by segment ID
	Put(segmentUsers *ftproto.SegmentUsers) error

	// Delete a feature flag
	Delete(id string)

	// Delete all the feature flags
	DeleteAll() error
}

type segmentUsersCache struct {
	cache Cache
}

func NewSegmentUsersCache(c Cache) SegmentUsersCache {
	return &segmentUsersCache{cache: c}
}

func (c *segmentUsersCache) Get(segmentID string) (*ftproto.SegmentUsers, error) {
	key := fmt.Sprintf("%s:%s", segmentUsersPrefix, segmentID)
	value, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}
	buffer, err := Bytes(value)
	if err != nil {
		return nil, ErrInvalidType
	}
	segmentUsers := &ftproto.SegmentUsers{}
	err = proto.Unmarshal(buffer, segmentUsers)
	if err != nil {
		return nil, ErrFailedToUnmarshalProto
	}
	return segmentUsers, nil
}

func (c *segmentUsersCache) GetSegmentIDs() ([]string, error) {
	keyPrefix := fmt.Sprintf("%s:", segmentUsersPrefix)
	keys, err := c.cache.Scan(keyPrefix)
	if err != nil {
		return nil, err
	}
	segIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		// Remove the prefix from the key
		segmentID := strings.Replace(key, keyPrefix, "", 1)
		// Get the segment user
		segment, err := c.Get(segmentID)
		if err != nil {
			return nil, err
		}
		// Append the segment ID
		segIDs = append(segIDs, segment.SegmentId)
	}
	return segIDs, nil
}

func (c *segmentUsersCache) Put(segmentUsers *ftproto.SegmentUsers) error {
	buffer, err := proto.Marshal(segmentUsers)
	if err != nil {
		return ErrFailedToMarshalProto
	}
	if buffer == nil {
		return ErrProtoMessageNil
	}
	key := fmt.Sprintf("%s:%s", segmentUsersPrefix, segmentUsers.SegmentId)
	return c.cache.Put(key, buffer, segmentUsersTTL)
}

func (c *segmentUsersCache) Delete(id string) {
	c.cache.Delete(id)
}

func (c *segmentUsersCache) DeleteAll() error {
	keyPrefix := fmt.Sprintf("%s:", segmentUsersPrefix)
	keys, err := c.cache.Scan(keyPrefix)
	if err != nil {
		return err
	}
	for _, key := range keys {
		c.cache.Delete(key)
	}
	return nil
}
