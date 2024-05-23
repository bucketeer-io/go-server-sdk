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
	"time"

	ftproto "github.com/bucketeer-io/bucketeer/proto/feature"
	"google.golang.org/protobuf/proto"
)

const (
	segmentUsersFlagsPrefix   = "bucketeer_segment_users"
	segmentUsersFlagsTTL      = time.Duration(0)
	segmentUserRequestedAtKey = "requested_at"
)

type SegmentUsersCache interface {
	// Get the segment users by segment ID
	GetSegmentUsers(segmentID string) (*ftproto.SegmentUsers, error)

	// Save the segment users by segment ID
	PutSegmentUsers(segmentUsers *ftproto.SegmentUsers) error

	// Get the last request timestamp returned from the server
	// This timestamp is used when requesting the latest cache from the server.
	GetRequestedAt() (int64, error)

	// Save the last request timestamp
	PutRequestedAt(timestamp int64) error
}

type segmentUsersCache struct {
	cache Cache
}

func NewSegmentUsersCache(c Cache) SegmentUsersCache {
	return &segmentUsersCache{cache: c}
}

func (c *segmentUsersCache) GetSegmentUsers(segmentID string) (*ftproto.SegmentUsers, error) {
	key := fmt.Sprintf("%s:%s", segmentUsersFlagsPrefix, segmentID)
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

func (c *segmentUsersCache) PutSegmentUsers(segmentUsers *ftproto.SegmentUsers) error {
	buffer, err := proto.Marshal(segmentUsers)
	if err != nil {
		return ErrFailedToMarshalProto
	}
	if buffer == nil {
		return ErrProtoMessageNil
	}
	key := fmt.Sprintf("%s:%s", segmentUsersFlagsPrefix, segmentUsers.SegmentId)
	return c.cache.Put(key, buffer, segmentUsersFlagsTTL)
}

func (c *segmentUsersCache) GetRequestedAt() (int64, error) {
	key := fmt.Sprintf("%s:%s", segmentUsersFlagsPrefix, segmentUserRequestedAtKey)
	value, err := c.cache.Get(key)
	if err != nil {
		return 0, err
	}
	v, err := Int64(value)
	if err != nil {
		return 0, ErrInvalidType
	}
	return v, nil
}

func (c *segmentUsersCache) PutRequestedAt(timestamp int64) error {
	key := fmt.Sprintf("%s:%s", segmentUsersFlagsPrefix, segmentUserRequestedAtKey)
	return c.cache.Put(key, timestamp, featureFlagsCacheTTL)
}
