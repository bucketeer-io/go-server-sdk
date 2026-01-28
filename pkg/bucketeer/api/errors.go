package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/retry"
)

// isRetryableStatusCode checks if an HTTP status code should trigger a retry.
func isRetryableStatusCode(code int) bool {
	switch code {
	case http.StatusTooManyRequests: // 429 - Rate limit
		return true
	case 499: // Client closed request (nginx)
		return true
	case http.StatusInternalServerError: // 500 - Internal server error (can be transient)
		return true
	case http.StatusBadGateway: // 502 - Bad gateway
		return true
	case http.StatusServiceUnavailable: // 503 - Service unavailable
		return true
	case http.StatusGatewayTimeout: // 504 - Gateway timeout
		return true
	default:
		return false
	}
}

// isRetryableError checks if an error should trigger a retry.
// This extends retry.IsRetryable with HTTP status code awareness.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// First, check if it's already a RetryableError from the retry package
	if retry.IsRetryable(err) {
		return true
	}

	// Check for HTTP status code errors (our status type)
	code, ok := GetStatusCode(err)
	if ok && code != http.StatusOK {
		return isRetryableStatusCode(code)
	}

	// Check for context deadline/cancellation
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for net.Error with Timeout() method
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Check for net.OpError (connection errors)
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check for url.Error (wraps network errors from http.Client)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

// parseRetryAfter parses the Retry-After header value.
// Returns 0 if not present or invalid.
// Supports both seconds (integer) and HTTP-date formats.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}

	// Try parsing as seconds (integer)
	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date
	if t, err := http.ParseTime(header); err == nil {
		duration := time.Until(t)
		if duration > 0 {
			return duration
		}
	}

	return 0
}
