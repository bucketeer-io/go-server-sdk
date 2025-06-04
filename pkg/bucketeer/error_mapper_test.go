package bucketeer

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/api"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/cache"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/evaluator"
	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
)

func TestMapErrorToReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		err               error
		isLocalEvaluation bool
		featureID         string
		expected          model.ReasonType
	}{
		{
			name:              "nil error returns default",
			err:               nil,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonDefault,
		},
		{
			name:              "local evaluation - evaluation not found",
			err:               evaluator.ErrEvaluationNotFound,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorFlagNotFound,
		},
		{
			name:              "local evaluation - cache not found",
			err:               cache.ErrNotFound,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorCacheNotFound,
		},
		{
			name:              "local evaluation - other error",
			err:               errors.New("some error"),
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "remote evaluation - 404 not found",
			err:               api.NewErrStatus(http.StatusNotFound),
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorFlagNotFound,
		},
		{
			name:              "remote evaluation - 500 server error",
			err:               api.NewErrStatus(http.StatusInternalServerError),
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "remote evaluation - 503 service unavailable",
			err:               api.NewErrStatus(http.StatusServiceUnavailable),
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "validation - invalid user with custom error",
			err:               ErrInvalidUser,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorUserIDNotSpecified,
		},
		{
			name:              "validation - invalid user with message",
			err:               errors.New("invalid user"),
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "validation - user ID is empty with custom error",
			err:               ErrUserIDEmpty,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorUserIDNotSpecified,
		},
		{
			name:              "validation - featureID is empty with custom error",
			err:               ErrFeatureIDEmpty,
			isLocalEvaluation: true,
			featureID:         "",
			expected:          model.ReasonErrorFeatureFlagIDNotSpecified,
		},
		{
			name:              "validation - cache not ready",
			err:               ErrCacheNotReady,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorCacheNotFound,
		},
		{
			name:              "validation - unsupported type",
			err:               ErrUnsupportedType,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorWrongType,
		},
		{
			name:              "local evaluation - cache failed to marshal proto",
			err:               cache.ErrFailedToMarshalProto,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorCacheNotFound,
		},
		{
			name:              "local evaluation - cache proto message nil",
			err:               cache.ErrProtoMessageNil,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorCacheNotFound,
		},
		{
			name:              "local evaluation - cache failed to unmarshal proto",
			err:               cache.ErrFailedToUnmarshalProto,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorCacheNotFound,
		},
		{
			name:              "local evaluation - cache invalid type",
			err:               cache.ErrInvalidType,
			isLocalEvaluation: true,
			featureID:         "test-feature",
			expected:          model.ReasonErrorCacheNotFound,
		},
		{
			name:              "response validation - nil response",
			err:               ErrResponseNil,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "response validation - nil evaluation",
			err:               ErrResponseEvaluationNil,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "response validation - empty variation value",
			err:               ErrResponseVariationValueEmpty,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "response validation - feature ID mismatch",
			err:               ErrResponseFeatureIDMismatch,
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
		{
			name:              "unknown error",
			err:               errors.New("unknown error"),
			isLocalEvaluation: false,
			featureID:         "test-feature",
			expected:          model.ReasonErrorException,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapErrorToReason(tt.err, tt.isLocalEvaluation, tt.featureID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

