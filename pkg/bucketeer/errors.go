package bucketeer

import "errors"

// Validation errors
var (
	ErrUserIDEmpty     = errors.New("user ID is empty")
	ErrInvalidUser     = errors.New("invalid user")
	ErrFeatureIDEmpty  = errors.New("featureID is empty")
	ErrCacheNotReady   = errors.New("cache is not ready")
	ErrUnsupportedType = errors.New("unsupported variation type")
)

// Response validation errors
var (
	ErrResponseNil                 = errors.New("invalid get evaluation response: res is nil")
	ErrResponseEvaluationNil       = errors.New("invalid get evaluation response: evaluation is nil")
	ErrResponseVariationValueEmpty = errors.New("invalid get evaluation response: variation value is empty")
	ErrResponseFeatureIDMismatch   = errors.New("invalid get evaluation response: feature id doesn't match")
)
