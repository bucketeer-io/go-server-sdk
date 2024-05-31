// Code generated by MockGen. DO NOT EDIT.
// Source: segment_user_processor.go

// Package processor is a generated GoMock package.
package processor

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockSegmentUserProcessor is a mock of SegmentUserProcessor interface.
type MockSegmentUserProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockSegmentUserProcessorMockRecorder
}

// MockSegmentUserProcessorMockRecorder is the mock recorder for MockSegmentUserProcessor.
type MockSegmentUserProcessorMockRecorder struct {
	mock *MockSegmentUserProcessor
}

// NewMockSegmentUserProcessor creates a new mock instance.
func NewMockSegmentUserProcessor(ctrl *gomock.Controller) *MockSegmentUserProcessor {
	mock := &MockSegmentUserProcessor{ctrl: ctrl}
	mock.recorder = &MockSegmentUserProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSegmentUserProcessor) EXPECT() *MockSegmentUserProcessorMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockSegmentUserProcessor) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockSegmentUserProcessorMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSegmentUserProcessor)(nil).Close))
}

// Run mocks base method.
func (m *MockSegmentUserProcessor) Run() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run")
}

// Run indicates an expected call of Run.
func (mr *MockSegmentUserProcessorMockRecorder) Run() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockSegmentUserProcessor)(nil).Run))
}