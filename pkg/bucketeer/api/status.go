package api

import (
	"fmt"
	"net/http"
	"time"
)

type status struct {
	code       int
	retryAfter time.Duration
}

type errStatus interface {
	c() int
}

type errStatusWithRetryAfter interface {
	errStatus
	ra() time.Duration
}

func NewErrStatus(code int) error {
	s := &status{
		code: code,
	}
	return s
}

// newErrStatusWithRetryAfter creates a status error with Retry-After duration.
func newErrStatusWithRetryAfter(code int, retryAfter time.Duration) error {
	s := &status{
		code:       code,
		retryAfter: retryAfter,
	}
	return s
}

func (s *status) Error() string {
	return fmt.Sprintf("bucketeer/api: send HTTP request failed: %d", s.code)
}

func (s *status) c() int {
	return s.code
}

func (s *status) ra() time.Duration {
	return s.retryAfter
}

func GetStatusCode(err error) (int, bool) {
	if err == nil {
		return http.StatusOK, true
	}
	s, ok := err.(errStatus)
	if !ok {
		return 0, false
	}
	return s.c(), true
}

// getRetryAfter extracts the Retry-After duration from a status error.
// Returns 0 if not present or not a status error.
func getRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}
	s, ok := err.(errStatusWithRetryAfter)
	if !ok {
		return 0
	}
	return s.ra()
}
