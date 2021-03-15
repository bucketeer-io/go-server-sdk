// Code generated by MockGen. DO NOT EDIT.
// Source: log.go

// Package log is a generated GoMock package.
package log

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockBaseLogger is a mock of BaseLogger interface
type MockBaseLogger struct {
	ctrl     *gomock.Controller
	recorder *MockBaseLoggerMockRecorder
}

// MockBaseLoggerMockRecorder is the mock recorder for MockBaseLogger
type MockBaseLoggerMockRecorder struct {
	mock *MockBaseLogger
}

// NewMockBaseLogger creates a new mock instance
func NewMockBaseLogger(ctrl *gomock.Controller) *MockBaseLogger {
	mock := &MockBaseLogger{ctrl: ctrl}
	mock.recorder = &MockBaseLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBaseLogger) EXPECT() *MockBaseLoggerMockRecorder {
	return m.recorder
}

// Print mocks base method
func (m *MockBaseLogger) Print(values ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range values {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Print", varargs...)
}

// Print indicates an expected call of Print
func (mr *MockBaseLoggerMockRecorder) Print(values ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Print", reflect.TypeOf((*MockBaseLogger)(nil).Print), values...)
}

// Printf mocks base method
func (m *MockBaseLogger) Printf(format string, values ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{format}
	for _, a := range values {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Printf", varargs...)
}

// Printf indicates an expected call of Printf
func (mr *MockBaseLoggerMockRecorder) Printf(format interface{}, values ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{format}, values...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Printf", reflect.TypeOf((*MockBaseLogger)(nil).Printf), varargs...)
}
