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
	segmentUsersFlagsKind = "bucketeer_segment_users"
	segmentUsersFlagsTTL  = time.Duration(0)
)

type SegmentUsersCache interface {
	Get(segmentID string) (*ftproto.SegmentUsers, error)
	Put(segmentUsers *ftproto.SegmentUsers) error
}

type segmentUsersCache struct {
	cache Cache
}

func NewSegmentUsersCache(c Cache) SegmentUsersCache {
	return &segmentUsersCache{cache: c}
}

func (c *segmentUsersCache) Get(segmentID string) (*ftproto.SegmentUsers, error) {
	key := c.newKey(segmentID)
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

func (c *segmentUsersCache) Put(segmentUsers *ftproto.SegmentUsers) error {
	buffer, err := proto.Marshal(segmentUsers)
	if err != nil {
		return ErrFailedToMarshalProto
	}
	if buffer == nil {
		return ErrProtoMessageNil
	}
	key := c.newKey(segmentUsers.SegmentId)
	return c.cache.Put(key, buffer, segmentUsersFlagsTTL)
}

func (c *segmentUsersCache) newKey(segmentID string) string {
	return fmt.Sprintf("%s:%s", segmentUsersFlagsKind, segmentID)
}
