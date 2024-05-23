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
	featureFlagsIDKey    = "id"
	featureFlagPrefix    = "bucketeer_feature_flag"
	featureFlagsPrefix   = "bucketeer_feature_flags"
	featureFlagsCacheTTL = time.Duration(0)
	requestedAtKey       = "requested_at"
)

type FeaturesCache interface {
	// Get the feature flag by ID
	GetFeatureFlag(id string) (*ftproto.Feature, error)

	// Save a feature flag
	PutFeatureFlag(feature *ftproto.Feature) error

	// Get the feature flags ID.
	// This ID represents a combination of all the features
	// saved in the cache. If the cache changes, the ID will change, too.
	// This ID is used when requesting the latest cache from the server.
	GetFeatureFlagsID() (string, error)

	// Save the feature flags ID
	PutFeatureFlagsID(id string) error

	// Get the last request timestamp returned from the server
	// This timestamp is used when requesting the latest cache from the server.
	GetRequestedAt() (int64, error)

	// Save the last request timestamp
	PutRequestedAt(timestamp int64) error
}

type featuresCache struct {
	cache Cache
}

func NewFeaturesCache(c Cache) FeaturesCache {
	return &featuresCache{cache: c}
}

func (c *featuresCache) GetFeatureFlag(id string) (*ftproto.Feature, error) {
	key := fmt.Sprintf("%s:%s", featureFlagPrefix, id)
	value, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}
	buffer, err := Bytes(value)
	if err != nil {
		return nil, ErrInvalidType
	}
	feature := &ftproto.Feature{}
	err = proto.Unmarshal(buffer, feature)
	if err != nil {
		return nil, ErrFailedToUnmarshalProto
	}
	return feature, nil
}

func (c *featuresCache) PutFeatureFlag(feature *ftproto.Feature) error {
	buffer, err := proto.Marshal(feature)
	if err != nil {
		return ErrFailedToMarshalProto
	}
	if buffer == nil {
		return ErrProtoMessageNil
	}
	key := fmt.Sprintf("%s:%s", featureFlagPrefix, feature.Id)
	return c.cache.Put(key, buffer, featureFlagsCacheTTL)
}

func (c *featuresCache) GetFeatureFlagsID() (string, error) {
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, featureFlagsIDKey)
	value, err := c.cache.Get(key)
	if err != nil {
		return "", err
	}
	v, err := String(value)
	if err != nil {
		return "", ErrInvalidType
	}
	return v, nil
}

func (c *featuresCache) PutFeatureFlagsID(id string) error {
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, featureFlagsIDKey)
	return c.cache.Put(key, id, featureFlagsCacheTTL)
}

func (c *featuresCache) GetRequestedAt() (int64, error) {
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, requestedAtKey)
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

func (c *featuresCache) PutRequestedAt(timestamp int64) error {
	key := fmt.Sprintf("%s:%s", featureFlagsPrefix, requestedAtKey)
	return c.cache.Put(key, timestamp, featureFlagsCacheTTL)
}
