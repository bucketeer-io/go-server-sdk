// Code generated by MockGen. DO NOT EDIT.
// Source: features.go
//
// Generated by this command:
//
//	mockgen -source=features.go -package=mock -destination=./mock/features.go
//

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	feature "github.com/bucketeer-io/bucketeer/proto/feature"
	gomock "go.uber.org/mock/gomock"
)

// MockFeaturesCache is a mock of FeaturesCache interface.
type MockFeaturesCache struct {
	ctrl     *gomock.Controller
	recorder *MockFeaturesCacheMockRecorder
}

// MockFeaturesCacheMockRecorder is the mock recorder for MockFeaturesCache.
type MockFeaturesCacheMockRecorder struct {
	mock *MockFeaturesCache
}

// NewMockFeaturesCache creates a new mock instance.
func NewMockFeaturesCache(ctrl *gomock.Controller) *MockFeaturesCache {
	mock := &MockFeaturesCache{ctrl: ctrl}
	mock.recorder = &MockFeaturesCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeaturesCache) EXPECT() *MockFeaturesCacheMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockFeaturesCache) Delete(id string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Delete", id)
}

// Delete indicates an expected call of Delete.
func (mr *MockFeaturesCacheMockRecorder) Delete(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockFeaturesCache)(nil).Delete), id)
}

// DeleteAll mocks base method.
func (m *MockFeaturesCache) DeleteAll() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAll")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAll indicates an expected call of DeleteAll.
func (mr *MockFeaturesCacheMockRecorder) DeleteAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockFeaturesCache)(nil).DeleteAll))
}

// Get mocks base method.
func (m *MockFeaturesCache) Get(id string) (*feature.Feature, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", id)
	ret0, _ := ret[0].(*feature.Feature)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockFeaturesCacheMockRecorder) Get(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockFeaturesCache)(nil).Get), id)
}

// Put mocks base method.
func (m *MockFeaturesCache) Put(feature *feature.Feature) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", feature)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockFeaturesCacheMockRecorder) Put(feature any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockFeaturesCache)(nil).Put), feature)
}
