// Code generated by MockGen. DO NOT EDIT.
// Source: sdk.go

// Package bucketeer is a generated GoMock package.
package bucketeer

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	model "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	user "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

// MockSDK is a mock of SDK interface.
type MockSDK struct {
	ctrl     *gomock.Controller
	recorder *MockSDKMockRecorder
}

// MockSDKMockRecorder is the mock recorder for MockSDK.
type MockSDKMockRecorder struct {
	mock *MockSDK
}

// NewMockSDK creates a new mock instance.
func NewMockSDK(ctrl *gomock.Controller) *MockSDK {
	mock := &MockSDK{ctrl: ctrl}
	mock.recorder = &MockSDKMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSDK) EXPECT() *MockSDKMockRecorder {
	return m.recorder
}

// BoolVariation mocks base method.
func (m *MockSDK) BoolVariation(ctx context.Context, user *user.User, featureID string, defaultValue bool) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BoolVariation", ctx, user, featureID, defaultValue)
	ret0, _ := ret[0].(bool)
	return ret0
}

// BoolVariation indicates an expected call of BoolVariation.
func (mr *MockSDKMockRecorder) BoolVariation(ctx, user, featureID, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BoolVariation", reflect.TypeOf((*MockSDK)(nil).BoolVariation), ctx, user, featureID, defaultValue)
}

// Close mocks base method.
func (m *MockSDK) Close(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockSDKMockRecorder) Close(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockSDK)(nil).Close), ctx)
}

// Float64Variation mocks base method.
func (m *MockSDK) Float64Variation(ctx context.Context, user *user.User, featureID string, defaultValue float64) float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Float64Variation", ctx, user, featureID, defaultValue)
	ret0, _ := ret[0].(float64)
	return ret0
}

// Float64Variation indicates an expected call of Float64Variation.
func (mr *MockSDKMockRecorder) Float64Variation(ctx, user, featureID, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Float64Variation", reflect.TypeOf((*MockSDK)(nil).Float64Variation), ctx, user, featureID, defaultValue)
}

// Int64Variation mocks base method.
func (m *MockSDK) Int64Variation(ctx context.Context, user *user.User, featureID string, defaultValue int64) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Int64Variation", ctx, user, featureID, defaultValue)
	ret0, _ := ret[0].(int64)
	return ret0
}

// Int64Variation indicates an expected call of Int64Variation.
func (mr *MockSDKMockRecorder) Int64Variation(ctx, user, featureID, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Int64Variation", reflect.TypeOf((*MockSDK)(nil).Int64Variation), ctx, user, featureID, defaultValue)
}

// IntVariation mocks base method.
func (m *MockSDK) IntVariation(ctx context.Context, user *user.User, featureID string, defaultValue int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IntVariation", ctx, user, featureID, defaultValue)
	ret0, _ := ret[0].(int)
	return ret0
}

// IntVariation indicates an expected call of IntVariation.
func (mr *MockSDKMockRecorder) IntVariation(ctx, user, featureID, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IntVariation", reflect.TypeOf((*MockSDK)(nil).IntVariation), ctx, user, featureID, defaultValue)
}

// JSONVariation mocks base method.
func (m *MockSDK) JSONVariation(ctx context.Context, user *user.User, featureID string, dst interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "JSONVariation", ctx, user, featureID, dst)
}

// JSONVariation indicates an expected call of JSONVariation.
func (mr *MockSDKMockRecorder) JSONVariation(ctx, user, featureID, dst interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JSONVariation", reflect.TypeOf((*MockSDK)(nil).JSONVariation), ctx, user, featureID, dst)
}

// StringVariation mocks base method.
func (m *MockSDK) StringVariation(ctx context.Context, user *user.User, featureID, defaultValue string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StringVariation", ctx, user, featureID, defaultValue)
	ret0, _ := ret[0].(string)
	return ret0
}

// StringVariation indicates an expected call of StringVariation.
func (mr *MockSDKMockRecorder) StringVariation(ctx, user, featureID, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StringVariation", reflect.TypeOf((*MockSDK)(nil).StringVariation), ctx, user, featureID, defaultValue)
}

// Track mocks base method.
func (m *MockSDK) Track(ctx context.Context, user *user.User, GoalID string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Track", ctx, user, GoalID)
}

// Track indicates an expected call of Track.
func (mr *MockSDKMockRecorder) Track(ctx, user, GoalID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Track", reflect.TypeOf((*MockSDK)(nil).Track), ctx, user, GoalID)
}

// TrackValue mocks base method.
func (m *MockSDK) TrackValue(ctx context.Context, user *user.User, GoalID string, value float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "TrackValue", ctx, user, GoalID, value)
}

// TrackValue indicates an expected call of TrackValue.
func (mr *MockSDKMockRecorder) TrackValue(ctx, user, GoalID, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TrackValue", reflect.TypeOf((*MockSDK)(nil).TrackValue), ctx, user, GoalID, value)
}

// getEvaluationLocally mocks base method.
func (m *MockSDK) getEvaluationLocally(ctx context.Context, user *user.User, featureID string) (*model.Evaluation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getEvaluationLocally", ctx, user, featureID)
	ret0, _ := ret[0].(*model.Evaluation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// getEvaluationLocally indicates an expected call of getEvaluationLocally.
func (mr *MockSDKMockRecorder) getEvaluationLocally(ctx, user, featureID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getEvaluationLocally", reflect.TypeOf((*MockSDK)(nil).getEvaluationLocally), ctx, user, featureID)
}

// getEvaluationRemotely mocks base method.
func (m *MockSDK) getEvaluationRemotely(ctx context.Context, user *user.User, featureID string) (*model.Evaluation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getEvaluationRemotely", ctx, user, featureID)
	ret0, _ := ret[0].(*model.Evaluation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// getEvaluationRemotely indicates an expected call of getEvaluationRemotely.
func (mr *MockSDKMockRecorder) getEvaluationRemotely(ctx, user, featureID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getEvaluationRemotely", reflect.TypeOf((*MockSDK)(nil).getEvaluationRemotely), ctx, user, featureID)
}
