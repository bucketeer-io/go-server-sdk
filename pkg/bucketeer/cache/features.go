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
	featureFlagsKind = "bucketeer_feature_flags"
	featureFlagsTTL  = time.Duration(0)
)

type FeaturesCache interface {
	Get(tag string) (*ftproto.Features, error)
	Put(tag string, features *ftproto.Features) error
}

type featuresCache struct {
	cache Cache
}

func NewFeaturesCache(c Cache) FeaturesCache {
	return &featuresCache{cache: c}
}

func (c *featuresCache) Get(tag string) (*ftproto.Features, error) {
	key := c.newKey(tag)
	value, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}
	buffer, err := Bytes(value)
	if err != nil {
		return nil, ErrInvalidType
	}
	features := &ftproto.Features{}
	err = proto.Unmarshal(buffer, features)
	if err != nil {
		return nil, ErrFailedToUnmarshalProto
	}
	return features, nil
}

func (c *featuresCache) Put(tag string, features *ftproto.Features) error {
	buffer, err := proto.Marshal(features)
	if err != nil {
		return ErrFailedToMarshalProto
	}
	if buffer == nil {
		return ErrProtoMessageNil
	}
	key := c.newKey(tag)
	return c.cache.Put(key, buffer, featureFlagsTTL)
}

func (c *featuresCache) newKey(tag string) string {
	if tag == "" {
		tag = "no_tag"
	}
	return fmt.Sprintf("%s:%s", featureFlagsKind, tag)
}
