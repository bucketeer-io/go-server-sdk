// Code generated by MockGen. DO NOT EDIT.
// Source: feature_flag_processor.go

// Package processor is a generated GoMock package.
package processor

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockFeatureFlagProcessor is a mock of FeatureFlagProcessor interface.
type MockFeatureFlagProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockFeatureFlagProcessorMockRecorder
}

// MockFeatureFlagProcessorMockRecorder is the mock recorder for MockFeatureFlagProcessor.
type MockFeatureFlagProcessorMockRecorder struct {
	mock *MockFeatureFlagProcessor
}

// NewMockFeatureFlagProcessor creates a new mock instance.
func NewMockFeatureFlagProcessor(ctrl *gomock.Controller) *MockFeatureFlagProcessor {
	mock := &MockFeatureFlagProcessor{ctrl: ctrl}
	mock.recorder = &MockFeatureFlagProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeatureFlagProcessor) EXPECT() *MockFeatureFlagProcessorMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockFeatureFlagProcessor) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockFeatureFlagProcessorMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockFeatureFlagProcessor)(nil).Close))
}

// Run mocks base method.
func (m *MockFeatureFlagProcessor) Run() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run")
}

// Run indicates an expected call of Run.
func (mr *MockFeatureFlagProcessorMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockFeatureFlagProcessor)(nil).Run))
}
