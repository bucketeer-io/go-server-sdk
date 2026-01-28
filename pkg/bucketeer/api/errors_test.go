package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/retry"
)

func TestIsRetryableStatusCode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		code     int
		expected bool
	}{
		{
			name:     "429 Too Many Requests is retryable",
			code:     http.StatusTooManyRequests,
			expected: true,
		},
		{
			name:     "499 Client Closed Request is retryable",
			code:     499,
			expected: true,
		},
		{
			name:     "500 Internal Server Error is retryable",
			code:     http.StatusInternalServerError,
			expected: true,
		},
		{
			name:     "502 Bad Gateway is retryable",
			code:     http.StatusBadGateway,
			expected: true,
		},
		{
			name:     "503 Service Unavailable is retryable",
			code:     http.StatusServiceUnavailable,
			expected: true,
		},
		{
			name:     "504 Gateway Timeout is retryable",
			code:     http.StatusGatewayTimeout,
			expected: true,
		},
		{
			name:     "400 Bad Request is not retryable",
			code:     http.StatusBadRequest,
			expected: false,
		},
		{
			name:     "401 Unauthorized is not retryable",
			code:     http.StatusUnauthorized,
			expected: false,
		},
		{
			name:     "403 Forbidden is not retryable",
			code:     http.StatusForbidden,
			expected: false,
		},
		{
			name:     "404 Not Found is not retryable",
			code:     http.StatusNotFound,
			expected: false,
		},
		{
			name:     "200 OK is not retryable",
			code:     http.StatusOK,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := isRetryableStatusCode(tc.code)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "RetryableError with Retryable=true",
			err:      &retry.RetryableError{Err: errors.New("test"), Retryable: true},
			expected: true,
		},
		{
			name:     "RetryableError with Retryable=false",
			err:      &retry.RetryableError{Err: errors.New("test"), Retryable: false},
			expected: false,
		},
		{
			name:     "status error with 500",
			err:      NewErrStatus(http.StatusInternalServerError),
			expected: true,
		},
		{
			name:     "status error with 502",
			err:      NewErrStatus(http.StatusBadGateway),
			expected: true,
		},
		{
			name:     "status error with 503",
			err:      NewErrStatus(http.StatusServiceUnavailable),
			expected: true,
		},
		{
			name:     "status error with 504",
			err:      NewErrStatus(http.StatusGatewayTimeout),
			expected: true,
		},
		{
			name:     "status error with 429",
			err:      NewErrStatus(http.StatusTooManyRequests),
			expected: true,
		},
		{
			name:     "status error with 400",
			err:      NewErrStatus(http.StatusBadRequest),
			expected: false,
		},
		{
			name:     "status error with 401",
			err:      NewErrStatus(http.StatusUnauthorized),
			expected: false,
		},
		{
			name:     "status error with 404",
			err:      NewErrStatus(http.StatusNotFound),
			expected: false,
		},
		{
			name:     "context.DeadlineExceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "wrapped context.DeadlineExceeded",
			err:      fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
			expected: true,
		},
		{
			name:     "net.OpError",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")},
			expected: true,
		},
		{
			name:     "wrapped net.OpError",
			err:      fmt.Errorf("wrapped: %w", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}),
			expected: true,
		},
		{
			name:     "url.Error",
			err:      &url.Error{Op: "Get", URL: "http://example.com", Err: errors.New("connection refused")},
			expected: true,
		},
		{
			name:     "wrapped url.Error",
			err:      fmt.Errorf("wrapped: %w", &url.Error{Op: "Get", URL: "http://example.com", Err: errors.New("connection refused")}),
			expected: true,
		},
		{
			name:     "net.DNSError",
			err:      &net.DNSError{Err: "no such host", Name: "example.com"},
			expected: true,
		},
		{
			name:     "wrapped net.DNSError",
			err:      fmt.Errorf("wrapped: %w", &net.DNSError{Err: "no such host", Name: "example.com"}),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "wrapped generic error",
			err:      fmt.Errorf("wrapped: %w", errors.New("some error")),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := isRetryableError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		header   string
		expected time.Duration
	}{
		{
			name:     "empty header",
			header:   "",
			expected: 0,
		},
		{
			name:     "integer seconds",
			header:   "120",
			expected: 120 * time.Second,
		},
		{
			name:     "single digit seconds",
			header:   "5",
			expected: 5 * time.Second,
		},
		{
			name:     "zero seconds",
			header:   "0",
			expected: 0,
		},
		{
			name:     "negative seconds",
			header:   "-10",
			expected: 0,
		},
		{
			name:     "invalid string",
			header:   "invalid",
			expected: 0,
		},
		{
			name:     "float value treated as invalid",
			header:   "1.5",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := parseRetryAfter(tc.header)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	t.Parallel()

	// Test with a future HTTP date
	futureTime := time.Now().Add(60 * time.Second)
	httpDate := futureTime.UTC().Format(http.TimeFormat)

	result := parseRetryAfter(httpDate)

	// Should return approximately 60 seconds (allow some tolerance for test execution time)
	assert.Greater(t, result, 55*time.Second)
	assert.Less(t, result, 65*time.Second)
}

func TestParseRetryAfterPastHTTPDate(t *testing.T) {
	t.Parallel()

	// Test with a past HTTP date
	pastTime := time.Now().Add(-60 * time.Second)
	httpDate := pastTime.UTC().Format(http.TimeFormat)

	result := parseRetryAfter(httpDate)

	// Should return 0 for past dates
	assert.Equal(t, time.Duration(0), result)
}
