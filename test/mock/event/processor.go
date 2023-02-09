// Code generated by MockGen. DO NOT EDIT.
// Source: processor.go

// Package event is a generated GoMock package.
package event

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"

	models "github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/models"
	user "github.com/ca-dp/bucketeer-go-server-sdk/pkg/bucketeer/user"
)

// MockProcessor is a mock of Processor interface.
type MockProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockProcessorMockRecorder
}

// MockProcessorMockRecorder is the mock recorder for MockProcessor.
type MockProcessorMockRecorder struct {
	mock *MockProcessor
}

// NewMockProcessor creates a new mock instance.
func NewMockProcessor(ctrl *gomock.Controller) *MockProcessor {
	mock := &MockProcessor{ctrl: ctrl}
	mock.recorder = &MockProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProcessor) EXPECT() *MockProcessorMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockProcessor) Close(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockProcessorMockRecorder) Close(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockProcessor)(nil).Close), ctx)
}

// PushDefaultEvaluationEvent mocks base method.
func (m *MockProcessor) PushDefaultEvaluationEvent(ctx context.Context, user *user.User, featureID string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushDefaultEvaluationEvent", ctx, user, featureID)
}

// PushDefaultEvaluationEvent indicates an expected call of PushDefaultEvaluationEvent.
func (mr *MockProcessorMockRecorder) PushDefaultEvaluationEvent(ctx, user, featureID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushDefaultEvaluationEvent", reflect.TypeOf((*MockProcessor)(nil).PushDefaultEvaluationEvent), ctx, user, featureID)
}

// PushEvaluationEvent mocks base method.
func (m *MockProcessor) PushEvaluationEvent(ctx context.Context, user *user.User, evaluation *models.Evaluation) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushEvaluationEvent", ctx, user, evaluation)
}

// PushEvaluationEvent indicates an expected call of PushEvaluationEvent.
func (mr *MockProcessorMockRecorder) PushEvaluationEvent(ctx, user, evaluation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushEvaluationEvent", reflect.TypeOf((*MockProcessor)(nil).PushEvaluationEvent), ctx, user, evaluation)
}

// PushGetEvaluationLatencyMetricsEvent mocks base method.
func (m *MockProcessor) PushGetEvaluationLatencyMetricsEvent(ctx context.Context, duration time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushGetEvaluationLatencyMetricsEvent", ctx, duration)
}

// PushGetEvaluationLatencyMetricsEvent indicates an expected call of PushGetEvaluationLatencyMetricsEvent.
func (mr *MockProcessorMockRecorder) PushGetEvaluationLatencyMetricsEvent(ctx, duration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushGetEvaluationLatencyMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushGetEvaluationLatencyMetricsEvent), ctx, duration)
}

// PushGetEvaluationSizeMetricsEvent mocks base method.
func (m *MockProcessor) PushGetEvaluationSizeMetricsEvent(ctx context.Context, sizeByte int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushGetEvaluationSizeMetricsEvent", ctx, sizeByte)
}

// PushGetEvaluationSizeMetricsEvent indicates an expected call of PushGetEvaluationSizeMetricsEvent.
func (mr *MockProcessorMockRecorder) PushGetEvaluationSizeMetricsEvent(ctx, sizeByte interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushGetEvaluationSizeMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushGetEvaluationSizeMetricsEvent), ctx, sizeByte)
}

// PushGoalEvent mocks base method.
func (m *MockProcessor) PushGoalEvent(ctx context.Context, user *user.User, GoalID string, value float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushGoalEvent", ctx, user, GoalID, value)
}

// PushGoalEvent indicates an expected call of PushGoalEvent.
func (mr *MockProcessorMockRecorder) PushGoalEvent(ctx, user, GoalID, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushGoalEvent", reflect.TypeOf((*MockProcessor)(nil).PushGoalEvent), ctx, user, GoalID, value)
}

// PushInternalErrorCountMetricsEvent mocks base method.
func (m *MockProcessor) PushInternalErrorCountMetricsEvent(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushInternalErrorCountMetricsEvent", ctx)
}

// PushInternalErrorCountMetricsEvent indicates an expected call of PushInternalErrorCountMetricsEvent.
func (mr *MockProcessorMockRecorder) PushInternalErrorCountMetricsEvent(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushInternalErrorCountMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushInternalErrorCountMetricsEvent), ctx)
}

// PushTimeoutErrorCountMetricsEvent mocks base method.
func (m *MockProcessor) PushTimeoutErrorCountMetricsEvent(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushTimeoutErrorCountMetricsEvent", ctx)
}

// PushTimeoutErrorCountMetricsEvent indicates an expected call of PushTimeoutErrorCountMetricsEvent.
func (mr *MockProcessorMockRecorder) PushTimeoutErrorCountMetricsEvent(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushTimeoutErrorCountMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushTimeoutErrorCountMetricsEvent), ctx)
}
