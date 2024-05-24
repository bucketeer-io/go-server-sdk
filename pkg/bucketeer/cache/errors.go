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

import "errors"

var (
	ErrFailedToMarshalProto   = errors.New("cache: failed to marshal proto message")
	ErrProtoMessageNil        = errors.New("cache: proto message is nil")
	ErrFailedToUnmarshalProto = errors.New("cache: failed to unmarshal proto message")
	ErrInvalidType            = errors.New("cache: invalid type")
	ErrNotFound               = errors.New("cache: not found")
)
