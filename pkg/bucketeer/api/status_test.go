package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertPostgresError(t *testing.T) {
	t.Parallel()
	patterns := map[string]struct {
		input        error
		expectedCode int
		expectedBool bool
	}{
		"StatusInternalServerError": {
			input:        NewErrStatus(http.StatusInternalServerError),
			expectedCode: http.StatusInternalServerError,
			expectedBool: true,
		},
		"status ok": {
			input:        nil,
			expectedCode: http.StatusOK,
			expectedBool: true,
		},
		"unknown error": {
			input:        errors.New("error"),
			expectedCode: 0,
			expectedBool: false,
		},
	}
	for msg, p := range patterns {
		t.Run(msg, func(t *testing.T) {
			code, ok := GetStatusCode(p.input)
			assert.Equal(t, p.expectedCode, code)
			assert.Equal(t, p.expectedBool, ok)
		})
	}
}
