package api

import (
	"fmt"
	"net/http"
)

type status struct {
	code int
}

type errStatus interface {
	c() int
}

func NewErrStatus(code int) error {
	s := &status{
		code: code,
	}
	return s
}

func (s *status) Error() string {
	return fmt.Sprintf("bucketeer/api: send HTTP request failed: %d", s.code)
}

func (s *status) c() int {
	return s.code
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
