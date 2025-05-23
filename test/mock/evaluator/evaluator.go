// Code generated by MockGen. DO NOT EDIT.
// Source: evaluator.go
//
// Generated by this command:
//
//	mockgen -source=evaluator.go -package=evaluator -destination=../../../test/mock/evaluator/evaluator.go
//

// Package evaluator is a generated GoMock package.
package evaluator

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	model "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/model"
	user "github.com/bucketeer-io/go-server-sdk/pkg/bucketeer/user"
)

// MockEvaluateLocally is a mock of EvaluateLocally interface.
type MockEvaluateLocally struct {
	ctrl     *gomock.Controller
	recorder *MockEvaluateLocallyMockRecorder
}

// MockEvaluateLocallyMockRecorder is the mock recorder for MockEvaluateLocally.
type MockEvaluateLocallyMockRecorder struct {
	mock *MockEvaluateLocally
}

// NewMockEvaluateLocally creates a new mock instance.
func NewMockEvaluateLocally(ctrl *gomock.Controller) *MockEvaluateLocally {
	mock := &MockEvaluateLocally{ctrl: ctrl}
	mock.recorder = &MockEvaluateLocallyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEvaluateLocally) EXPECT() *MockEvaluateLocallyMockRecorder {
	return m.recorder
}

// Evaluate mocks base method.
func (m *MockEvaluateLocally) Evaluate(user *user.User, featureID string) (*model.Evaluation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Evaluate", user, featureID)
	ret0, _ := ret[0].(*model.Evaluation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Evaluate indicates an expected call of Evaluate.
func (mr *MockEvaluateLocallyMockRecorder) Evaluate(user, featureID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Evaluate", reflect.TypeOf((*MockEvaluateLocally)(nil).Evaluate), user, featureID)
}
