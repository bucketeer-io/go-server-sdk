// Code generated by MockGen. DO NOT EDIT.
// Source: processor.go

// Package event is a generated GoMock package.
package event

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"

	model "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	user "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
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

// PushErrorEvent mocks base method.
func (m *MockProcessor) PushErrorEvent(ctx context.Context, err error, api model.APIID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushErrorEvent", ctx, err, api)
}

// PushErrorEvent indicates an expected call of PushErrorEvent.
func (mr *MockProcessorMockRecorder) PushErrorEvent(ctx, err, api interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushErrorEvent", reflect.TypeOf((*MockProcessor)(nil).PushErrorEvent), ctx, err, api)
}

// PushEvaluationEvent mocks base method.
func (m *MockProcessor) PushEvaluationEvent(ctx context.Context, user *user.User, evaluation *model.Evaluation) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushEvaluationEvent", ctx, user, evaluation)
}

// PushEvaluationEvent indicates an expected call of PushEvaluationEvent.
func (mr *MockProcessorMockRecorder) PushEvaluationEvent(ctx, user, evaluation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushEvaluationEvent", reflect.TypeOf((*MockProcessor)(nil).PushEvaluationEvent), ctx, user, evaluation)
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

// PushLatencyMetricsEvent mocks base method.
func (m *MockProcessor) PushLatencyMetricsEvent(ctx context.Context, duration time.Duration, api model.APIID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushLatencyMetricsEvent", ctx, duration, api)
}

// PushLatencyMetricsEvent indicates an expected call of PushLatencyMetricsEvent.
func (mr *MockProcessorMockRecorder) PushLatencyMetricsEvent(ctx, duration, api interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushLatencyMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushLatencyMetricsEvent), ctx, duration, api)
}

// PushSizeMetricsEvent mocks base method.
func (m *MockProcessor) PushSizeMetricsEvent(ctx context.Context, sizeByte int, api model.APIID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PushSizeMetricsEvent", ctx, sizeByte, api)
}

// PushSizeMetricsEvent indicates an expected call of PushSizeMetricsEvent.
func (mr *MockProcessorMockRecorder) PushSizeMetricsEvent(ctx, sizeByte, api interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushSizeMetricsEvent", reflect.TypeOf((*MockProcessor)(nil).PushSizeMetricsEvent), ctx, sizeByte, api)
}
