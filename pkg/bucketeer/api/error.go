package api

import (
	"fmt"
)

type status struct {
	code int
}

type ErrStatus interface {
	GetStatusCode() int
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

func (s *status) GetStatusCode() int {
	return s.code
}

func ConvertToErrStatus(err error) (ErrStatus, bool) {
	s, ok := err.(ErrStatus)
	if !ok {
		return nil, false
	}
	return s, true
}
