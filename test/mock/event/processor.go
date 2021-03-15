// Code generated by MockGen. DO NOT EDIT.
// Source: processor.go

// Package event is a generated GoMock package.
package event

import (
	context "context"
	feature "github.com/ca-dp/bucketeer-go-server-sdk/proto/feature"
	user "github.com/ca-dp/bucketeer-go-server-sdk/proto/user"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockProcessor is a mock of Processor interface
type MockProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockProcessorMockRecorder
}

// MockProcessorMockRecorder is the mock recorder for MockProcessor
type MockProcessorMockRecorder struct {
	mock *MockProcessor
}

// NewMockProcessor creates a new mock instance
func NewMockProcessor(ctrl *gomock.Controller) *MockProcessor {
	mock := &MockProcessor{ctrl: ctrl}
	mock.recorder = &MockProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProcessor) EXPECT() *MockProcessorMockRecorder {
	return m.recorder
}

// PushEvaluationEvent mocks base method
func (m *MockProcessor) PushEvaluationEvent(ctx context.Context, user *user.User, evaluation *feature.Evaluation) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushEvaluationEvent", ctx, user, evaluation)
}

// PushEvaluationEvent indicates an expected call of PushEvaluationEvent
func (mr *MockProcessorMockRecorder) PushEvaluationEvent(ctx, user, evaluation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushEvaluationEvent", reflect.TypeOf((*MockProcessor)(nil).PushEvaluationEvent), ctx, user, evaluation)
}

// PushDefaultEvaluationEvent mocks base method
func (m *MockProcessor) PushDefaultEvaluationEvent(ctx context.Context, user *user.User, featureID string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushDefaultEvaluationEvent", ctx, user, featureID)
}

// PushDefaultEvaluationEvent indicates an expected call of PushDefaultEvaluationEvent
func (mr *MockProcessorMockRecorder) PushDefaultEvaluationEvent(ctx, user, featureID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushDefaultEvaluationEvent", reflect.TypeOf((*MockProcessor)(nil).PushDefaultEvaluationEvent), ctx, user, featureID)
}

// PushGoalEvent mocks base method
func (m *MockProcessor) PushGoalEvent(ctx context.Context, user *user.User, goalID string, value float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushGoalEvent", ctx, user, goalID, value)
}

// PushGoalEvent indicates an expected call of PushGoalEvent
func (mr *MockProcessorMockRecorder) PushGoalEvent(ctx, user, goalID, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushGoalEvent", reflect.TypeOf((*MockProcessor)(nil).PushGoalEvent), ctx, user, goalID, value)
}

// PushGetEvaluationLatencyMetricsEvent mocks base method
func (m *MockProcessor) PushGetEvaluationLatencyMetricsEvent(ctx context.Context, duration time.Duration, tag string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushGetEvaluationLatencyMetricsEvent", ctx, duration, tag)
}

// PushGetEvaluationLatencyMetricsEvent indicates an expected call of PushGetEvaluationLatencyMetricsEvent
func (mr *MockProcessorMockRecorder) PushGetEvaluationLatencyMetricsEvent(ctx, duration, tag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushGetEvaluationLatencyMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushGetEvaluationLatencyMetricsEvent), ctx, duration, tag)
}

// PushGetEvaluationSizeMetricsEvent mocks base method
func (m *MockProcessor) PushGetEvaluationSizeMetricsEvent(ctx context.Context, sizeByte int, tag string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushGetEvaluationSizeMetricsEvent", ctx, sizeByte, tag)
}

// PushGetEvaluationSizeMetricsEvent indicates an expected call of PushGetEvaluationSizeMetricsEvent
func (mr *MockProcessorMockRecorder) PushGetEvaluationSizeMetricsEvent(ctx, sizeByte, tag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushGetEvaluationSizeMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushGetEvaluationSizeMetricsEvent), ctx, sizeByte, tag)
}

// PushTimeoutErrorCountMetricsEvent mocks base method
func (m *MockProcessor) PushTimeoutErrorCountMetricsEvent(ctx context.Context, tag string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushTimeoutErrorCountMetricsEvent", ctx, tag)
}

// PushTimeoutErrorCountMetricsEvent indicates an expected call of PushTimeoutErrorCountMetricsEvent
func (mr *MockProcessorMockRecorder) PushTimeoutErrorCountMetricsEvent(ctx, tag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushTimeoutErrorCountMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushTimeoutErrorCountMetricsEvent), ctx, tag)
}

// PushInternalErrorCountMetricsEvent mocks base method
func (m *MockProcessor) PushInternalErrorCountMetricsEvent(ctx context.Context, tag string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushInternalErrorCountMetricsEvent", ctx, tag)
}

// PushInternalErrorCountMetricsEvent indicates an expected call of PushInternalErrorCountMetricsEvent
func (mr *MockProcessorMockRecorder) PushInternalErrorCountMetricsEvent(ctx, tag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushInternalErrorCountMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushInternalErrorCountMetricsEvent), ctx, tag)
}

// Close mocks base method
func (m *MockProcessor) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockProcessorMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockProcessor)(nil).Close))
}