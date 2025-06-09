package bucketeer

import (
	"errors"
	"net/http"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/evaluator"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

func MapErrorToReason(err error, isLocalEvaluation bool, featureID string) model.ReasonType {
	if err == nil {
		return model.ReasonDefault
	}

	if errors.Is(err, user.ErrInvalidUser) || errors.Is(err, user.ErrUserIDEmpty) {
		return model.ReasonErrorUserIDNotSpecified
	}
	if errors.Is(err, ErrFeatureIDEmpty) {
		return model.ReasonErrorFeatureFlagIDNotSpecified
	}
	if errors.Is(err, ErrUnsupportedType) {
		return model.ReasonErrorWrongType
	}

	// Check for response validation errors
	if errors.Is(err, ErrResponseNil) ||
		errors.Is(err, ErrResponseEvaluationNil) ||
		errors.Is(err, ErrResponseVariationValueEmpty) ||
		errors.Is(err, ErrResponseFeatureIDMismatch) {
		return model.ReasonErrorException
	}

	if isLocalEvaluation {
		return mapLocalEvaluationError(err)
	}
	return mapRemoteEvaluationError(err)
}

func mapLocalEvaluationError(err error) model.ReasonType {
	if errors.Is(err, evaluator.ErrEvaluationNotFound) || errors.Is(err, cache.ErrNotFound) {
		return model.ReasonErrorFlagNotFound
	}
	if errors.Is(err, ErrCacheNotReady) || errors.Is(err, cache.ErrInvalidType) {
		return model.ReasonErrorCacheNotFound
	}
	if errors.Is(err, cache.ErrFailedToMarshalProto) ||
		errors.Is(err, cache.ErrProtoMessageNil) ||
		errors.Is(err, cache.ErrFailedToUnmarshalProto) {
		return model.ReasonErrorException
	}

	return model.ReasonErrorException
}

func mapRemoteEvaluationError(err error) model.ReasonType {
	// Only 404 Not Found gets a special error type
	if code, ok := api.GetStatusCode(err); ok && code == http.StatusNotFound {
		return model.ReasonErrorFlagNotFound
	}
	// All other errors (including non-HTTP errors) are treated as exceptions
	return model.ReasonErrorException
}
