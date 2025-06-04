package bucketeer

import (
	"errors"
	"net/http"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/evaluator"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

func MapErrorToReason(err error, isLocalEvaluation bool, featureID string) model.ReasonType {
	if err == nil {
		return model.ReasonDefault
	}

	if errors.Is(err, ErrInvalidUser) || errors.Is(err, ErrUserIDEmpty) {
		return model.ReasonErrorUserIDNotSpecified
	}
	if errors.Is(err, ErrFeatureIDEmpty) {
		return model.ReasonErrorFeatureFlagIDNotSpecified
	}
	if errors.Is(err, ErrUnsupportedType) {
		return model.ReasonErrorWrongType
	}

	// Check for response validation errors
	if errors.Is(err, ErrResponseNil) || errors.Is(err, ErrResponseEvaluationNil) ||
		errors.Is(err, ErrResponseVariationValueEmpty) || errors.Is(err, ErrResponseFeatureIDMismatch) {
		return model.ReasonErrorException
	}

	if isLocalEvaluation {
		return mapLocalEvaluationError(err)
	}
	return mapRemoteEvaluationError(err)
}

func mapLocalEvaluationError(err error) model.ReasonType {
	if errors.Is(err, evaluator.ErrEvaluationNotFound) {
		return model.ReasonErrorFlagNotFound
	}
	if errors.Is(err, cache.ErrNotFound) || errors.Is(err, ErrCacheNotReady) ||
		errors.Is(err, cache.ErrFailedToMarshalProto) || errors.Is(err, cache.ErrProtoMessageNil) ||
		errors.Is(err, cache.ErrFailedToUnmarshalProto) || errors.Is(err, cache.ErrInvalidType) {
		return model.ReasonErrorCacheNotFound
	}

	return model.ReasonErrorException
}

func mapRemoteEvaluationError(err error) model.ReasonType {
	// Check for HTTP status codes
	if code, ok := api.GetStatusCode(err); ok {
		switch code {
		case http.StatusNotFound:
			return model.ReasonErrorFlagNotFound
		case http.StatusBadRequest:
			return model.ReasonErrorException
		case http.StatusUnauthorized, http.StatusForbidden:
			return model.ReasonErrorException
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return model.ReasonErrorException
		default:
			return model.ReasonErrorException
		}
	}

	return model.ReasonErrorException
}
