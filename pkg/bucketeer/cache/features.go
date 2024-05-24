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
	featureFlagPrefix    = "bucketeer_feature_flag"
	featureFlagsCacheTTL = time.Duration(0)
)

type FeaturesCache interface {
	// Get the feature flag by ID
	Get(id string) (*ftproto.Feature, error)

	// Save a feature flag
	Put(feature *ftproto.Feature) error
}

type featuresCache struct {
	cache Cache
}

func NewFeaturesCache(c Cache) FeaturesCache {
	return &featuresCache{cache: c}
}

func (c *featuresCache) Get(id string) (*ftproto.Feature, error) {
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

func (c *featuresCache) Put(feature *ftproto.Feature) error {
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
