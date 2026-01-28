package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRetryable(t *testing.T) {
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
			err:      &RetryableError{Err: errors.New("test"), Retryable: true},
			expected: true,
		},
		{
			name:     "RetryableError with Retryable=false",
			err:      &RetryableError{Err: errors.New("test"), Retryable: false},
			expected: false,
		},
		{
			name:     "wrapped RetryableError with Retryable=true",
			err:      fmt.Errorf("wrapped: %w", &RetryableError{Err: errors.New("test"), Retryable: true}),
			expected: true,
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
			name:     "context.Canceled",
			err:      context.Canceled,
			expected: false,
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
			result := IsRetryable(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRetryableError(t *testing.T) {
	t.Parallel()

	t.Run("Error method returns wrapped error message", func(t *testing.T) {
		t.Parallel()
		innerErr := errors.New("inner error")
		re := &RetryableError{Err: innerErr, Retryable: true}
		assert.Equal(t, "inner error", re.Error())
	})

	t.Run("Unwrap returns inner error", func(t *testing.T) {
		t.Parallel()
		innerErr := errors.New("inner error")
		re := &RetryableError{Err: innerErr, Retryable: true}
		assert.Equal(t, innerErr, re.Unwrap())
	})

	t.Run("errors.Is works through RetryableError", func(t *testing.T) {
		t.Parallel()
		re := &RetryableError{Err: context.DeadlineExceeded, Retryable: true}
		assert.True(t, errors.Is(re, context.DeadlineExceeded))
	})
}

func TestCalculateBackoff(t *testing.T) {
	t.Parallel()

	t.Run("first attempt uses initial interval with jitter", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			InitialInterval: 1 * time.Second,
			MaxInterval:     10 * time.Second,
			Multiplier:      2.0,
		}

		// Run multiple times to account for jitter
		for i := 0; i < 100; i++ {
			backoff := calculateBackoff(0, cfg)
			// With ±25% jitter, backoff should be between 750ms and 1250ms
			assert.GreaterOrEqual(t, backoff, 750*time.Millisecond, "backoff should be >= 750ms")
			assert.LessOrEqual(t, backoff, 1250*time.Millisecond, "backoff should be <= 1250ms")
		}
	})

	t.Run("second attempt doubles with jitter", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			InitialInterval: 1 * time.Second,
			MaxInterval:     10 * time.Second,
			Multiplier:      2.0,
		}

		// Run multiple times to account for jitter
		for i := 0; i < 100; i++ {
			backoff := calculateBackoff(1, cfg)
			// 1s * 2^1 = 2s, with ±25% jitter: 1.5s to 2.5s
			assert.GreaterOrEqual(t, backoff, 1500*time.Millisecond, "backoff should be >= 1500ms")
			assert.LessOrEqual(t, backoff, 2500*time.Millisecond, "backoff should be <= 2500ms")
		}
	})

	t.Run("backoff is capped at MaxInterval", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			Multiplier:      2.0,
		}

		// Run multiple times to account for jitter
		for i := 0; i < 100; i++ {
			// Attempt 10: 1s * 2^10 = 1024s, but should be capped at 5s
			backoff := calculateBackoff(10, cfg)
			assert.LessOrEqual(t, backoff, 5*time.Second, "backoff should be capped at MaxInterval")
		}
	})

	t.Run("zero initial interval returns zero", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			InitialInterval: 0,
			MaxInterval:     10 * time.Second,
			Multiplier:      2.0,
		}

		backoff := calculateBackoff(0, cfg)
		assert.Equal(t, time.Duration(0), backoff)
	})

	t.Run("default multiplier is 2.0 when not set", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			InitialInterval: 1 * time.Second,
			MaxInterval:     10 * time.Second,
			Multiplier:      0, // Not set
		}

		// Run multiple times to account for jitter
		for i := 0; i < 100; i++ {
			backoff := calculateBackoff(1, cfg)
			// Should use default multiplier 2.0: 1s * 2^1 = 2s, with ±25% jitter: 1.5s to 2.5s
			assert.GreaterOrEqual(t, backoff, 1500*time.Millisecond)
			assert.LessOrEqual(t, backoff, 2500*time.Millisecond)
		}
	})

	t.Run("no max interval means no cap", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			InitialInterval: 1 * time.Second,
			MaxInterval:     0, // No cap
			Multiplier:      2.0,
		}

		// Attempt 3: 1s * 2^3 = 8s, with ±25% jitter: 6s to 10s
		for i := 0; i < 100; i++ {
			backoff := calculateBackoff(3, cfg)
			assert.GreaterOrEqual(t, backoff, 6*time.Second)
			assert.LessOrEqual(t, backoff, 10*time.Second)
		}
	})
}

func TestDo(t *testing.T) {
	t.Parallel()

	t.Run("successful on first attempt", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("successful after transient failures", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 3 {
				return &RetryableError{Err: errors.New("transient error"), Retryable: true}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("max retries exhausted", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("persistent error"), Retryable: true}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "max retries (3) exceeded")
		assert.Equal(t, 4, attempts) // 1 initial + 3 retries
	})

	t.Run("non-retryable error fails immediately", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		expectedErr := errors.New("non-retryable error")
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return expectedErr
		})

		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("RetryableError with Retryable=false fails immediately", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("not retryable"), Retryable: false}
		})

		require.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("context cancellation stops retries", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      10,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2.0,
		}

		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		err := Do(ctx, cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("transient error"), Retryable: true}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
		// Should have stopped early due to cancellation
		assert.Less(t, attempts, 10)
	})

	t.Run("deadline enforcement", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      10,
			InitialInterval: 100 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2.0,
			Deadline:        time.Now().Add(150 * time.Millisecond),
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("transient error"), Retryable: true}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "retry deadline exceeded")
		// Should have made at least 1 attempt, but not all 10
		assert.GreaterOrEqual(t, attempts, 1)
		assert.Less(t, attempts, 10)
	})

	t.Run("expired deadline fails immediately", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
			Deadline:        time.Now().Add(-1 * time.Second), // Already expired
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return nil
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "retry deadline exceeded")
		assert.Equal(t, 0, attempts)
	})

	t.Run("Retry-After is respected", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     1 * time.Second,
			Multiplier:      2.0,
		}

		attempts := 0
		start := time.Now()
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 2 {
				// Request 200ms retry delay
				return &RetryableError{
					Err:        errors.New("rate limited"),
					Retryable:  true,
					RetryAfter: 200 * time.Millisecond,
				}
			}
			return nil
		})

		elapsed := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
		// Should have waited at least 200ms (the Retry-After value)
		assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond)
	})

	t.Run("Retry-After is capped at MaxInterval", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond, // Cap at 100ms
			Multiplier:      2.0,
		}

		attempts := 0
		start := time.Now()
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 2 {
				// Request 500ms retry delay, but should be capped at 100ms
				return &RetryableError{
					Err:        errors.New("rate limited"),
					Retryable:  true,
					RetryAfter: 500 * time.Millisecond,
				}
			}
			return nil
		})

		elapsed := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
		// Should have waited around 100ms (the MaxInterval cap), not 500ms
		assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
		assert.Less(t, elapsed, 300*time.Millisecond) // Definitely less than 500ms
	})

	t.Run("zero max retries means no retries", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      0,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("transient error"), Retryable: true}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "max retries (0) exceeded")
		assert.Equal(t, 1, attempts) // Only the initial attempt
	})

	t.Run("network error is retried", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 3 {
				return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("url.Error is retried", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 3 {
				return &url.Error{Op: "Get", URL: "http://example.com", Err: errors.New("connection refused")}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("context.DeadlineExceeded is retried", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 3 {
				return context.DeadlineExceeded
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("context.Canceled is not retried", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return context.Canceled
		})

		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("DNS error is retried", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 3 {
				return &net.DNSError{Err: "no such host", Name: "example.com"}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})
}

func TestDoWithContextCancellation(t *testing.T) {
	t.Parallel()

	t.Run("context already canceled before first attempt", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      3,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     100 * time.Millisecond,
			Multiplier:      2.0,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		attempts := 0
		err := Do(ctx, cfg, func() error {
			attempts++
			return nil
		})

		require.Error(t, err)
		assert.Equal(t, 0, attempts)
	})

	t.Run("context canceled during backoff", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      10,
			InitialInterval: 500 * time.Millisecond, // Long backoff
			MaxInterval:     1 * time.Second,
			Multiplier:      2.0,
		}

		ctx, cancel := context.WithCancel(context.Background())

		attempts := 0
		go func() {
			time.Sleep(100 * time.Millisecond) // Cancel during backoff
			cancel()
		}()

		err := Do(ctx, cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("transient"), Retryable: true}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled during backoff")
		assert.Equal(t, 1, attempts) // Only first attempt before cancellation during backoff
	})
}

func TestDoWithDeadline(t *testing.T) {
	t.Parallel()

	t.Run("deadline allows multiple retries", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      10,
			InitialInterval: 10 * time.Millisecond,
			MaxInterval:     50 * time.Millisecond,
			Multiplier:      2.0,
			Deadline:        time.Now().Add(500 * time.Millisecond),
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 5 {
				return &RetryableError{Err: errors.New("transient"), Retryable: true}
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 5, attempts)
	})

	t.Run("deadline exceeded during backoff calculation", func(t *testing.T) {
		t.Parallel()
		cfg := Config{
			MaxRetries:      10,
			InitialInterval: 1 * time.Second, // Long backoff
			MaxInterval:     5 * time.Second,
			Multiplier:      2.0,
			Deadline:        time.Now().Add(100 * time.Millisecond), // Short deadline
		}

		attempts := 0
		err := Do(context.Background(), cfg, func() error {
			attempts++
			return &RetryableError{Err: errors.New("transient"), Retryable: true}
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "retry deadline exceeded")
		// Should have attempted at least once
		assert.GreaterOrEqual(t, attempts, 1)
	})
}
