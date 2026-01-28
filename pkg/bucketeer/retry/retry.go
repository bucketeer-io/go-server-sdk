package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/url"
	"time"
)

// Config holds retry configuration.
type Config struct {
	MaxRetries      int           // Maximum number of retry attempts
	InitialInterval time.Duration // Initial wait between retries
	MaxInterval     time.Duration // Maximum wait between retries
	Multiplier      float64       // Backoff multiplier
	Deadline        time.Time     // Must complete before this time (zero = no deadline)
}

// RetryableError wraps an error and indicates whether it's retryable.
// It can also carry a suggested retry delay from Retry-After headers.
type RetryableError struct {
	Err        error
	Retryable  bool
	RetryAfter time.Duration // Suggested delay from server (0 = use calculated backoff)
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error should be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for RetryableError wrapper
	var retryableErr *RetryableError
	if errors.As(err, &retryableErr) {
		return retryableErr.Retryable
	}

	// Check for context deadline/cancellation
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return false // Context canceled is not retryable
	}

	// Check for net.Error with Timeout() method
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true // All net.Error types are retryable (including timeouts)
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

// Do executes the function with retry logic.
// Returns the last error if all retries are exhausted.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation before attempting
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("context canceled: %w", lastErr)
			}
			return ctx.Err()
		default:
		}

		// Check deadline before attempting
		if !cfg.Deadline.IsZero() {
			remaining := time.Until(cfg.Deadline)
			if remaining <= 0 {
				if lastErr != nil {
					return fmt.Errorf("retry deadline exceeded: %w", lastErr)
				}
				return fmt.Errorf("retry deadline exceeded")
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			return err
		}

		// Don't sleep after last attempt
		if attempt == cfg.MaxRetries {
			break
		}

		// Calculate backoff (respect Retry-After if provided)
		backoff := calculateBackoff(attempt, cfg)
		var retryableErr *RetryableError
		if errors.As(err, &retryableErr) && retryableErr.RetryAfter > 0 {
			backoff = retryableErr.RetryAfter
			// Still cap at MaxInterval for safety
			if backoff > cfg.MaxInterval {
				backoff = cfg.MaxInterval
			}
		}

		// Ensure we don't exceed deadline
		if !cfg.Deadline.IsZero() {
			remaining := time.Until(cfg.Deadline)
			if backoff > remaining {
				// Not enough time for backoff + retry
				return fmt.Errorf("retry deadline exceeded: %w", lastErr)
			}
		}

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled during backoff: %w", lastErr)
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

// calculateBackoff returns the next backoff interval with jitter.
func calculateBackoff(attempt int, cfg Config) time.Duration {
	if cfg.InitialInterval <= 0 {
		return 0
	}

	// Exponential backoff: initialInterval * (multiplier ^ attempt)
	multiplier := cfg.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0 // Default multiplier
	}

	backoff := float64(cfg.InitialInterval) * math.Pow(multiplier, float64(attempt))

	// Add jitter (±25%) before capping
	jitter := backoff * 0.25 * (rand.Float64()*2 - 1)
	backoff += jitter

	// Cap at max interval (hard cap, applied after jitter)
	if cfg.MaxInterval > 0 && backoff > float64(cfg.MaxInterval) {
		backoff = float64(cfg.MaxInterval)
	}

	// Ensure backoff is not negative (edge case with jitter on small values)
	if backoff < 0 {
		backoff = float64(cfg.InitialInterval)
	}

	return time.Duration(backoff)
}
