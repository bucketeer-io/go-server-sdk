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
	"time"
)

type Cache interface {
	Getter
	Putter
	Deleter
	Scanner
}

type Getter interface {
	Get(key interface{}) (interface{}, error)
}

type Putter interface {
	Put(key interface{}, value interface{}, expiration time.Duration) error
}

type Deleter interface {
	Delete(key interface{})
}

type Scanner interface {
	Scan(keyPrefix string) ([]string, error)
}

func Bytes(value interface{}) ([]byte, error) {
	b, ok := value.([]byte)
	if !ok {
		return nil, ErrInvalidType
	}
	return b, nil
}

func String(value interface{}) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", ErrInvalidType
	}
	return s, nil
}

func Int64(value interface{}) (int64, error) {
	i, ok := value.(int64)
	if !ok {
		return 0, ErrInvalidType
	}
	return i, nil
}
