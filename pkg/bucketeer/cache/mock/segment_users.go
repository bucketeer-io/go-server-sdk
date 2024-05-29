// Code generated by MockGen. DO NOT EDIT.
// Source: segment_users.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	feature "github.com/bucketeer-io/bucketeer/proto/feature"
	gomock "go.uber.org/mock/gomock"
)

// MockSegmentUsersCache is a mock of SegmentUsersCache interface.
type MockSegmentUsersCache struct {
	ctrl     *gomock.Controller
	recorder *MockSegmentUsersCacheMockRecorder
}

// MockSegmentUsersCacheMockRecorder is the mock recorder for MockSegmentUsersCache.
type MockSegmentUsersCacheMockRecorder struct {
	mock *MockSegmentUsersCache
}

// NewMockSegmentUsersCache creates a new mock instance.
func NewMockSegmentUsersCache(ctrl *gomock.Controller) *MockSegmentUsersCache {
	mock := &MockSegmentUsersCache{ctrl: ctrl}
	mock.recorder = &MockSegmentUsersCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSegmentUsersCache) EXPECT() *MockSegmentUsersCacheMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockSegmentUsersCache) Delete(id string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Delete", id)
}

// Delete indicates an expected call of Delete.
func (mr *MockSegmentUsersCacheMockRecorder) Delete(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSegmentUsersCache)(nil).Delete), id)
}

// DeleteAll mocks base method.
func (m *MockSegmentUsersCache) DeleteAll() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAll")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAll indicates an expected call of DeleteAll.
func (mr *MockSegmentUsersCacheMockRecorder) DeleteAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockSegmentUsersCache)(nil).DeleteAll))
}

// Get mocks base method.
func (m *MockSegmentUsersCache) Get(segmentID string) (*feature.SegmentUsers, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", segmentID)
	ret0, _ := ret[0].(*feature.SegmentUsers)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSegmentUsersCacheMockRecorder) Get(segmentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSegmentUsersCache)(nil).Get), segmentID)
}

// GetSegmentIDs mocks base method.
func (m *MockSegmentUsersCache) GetSegmentIDs() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSegmentIDs")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSegmentIDs indicates an expected call of GetSegmentIDs.
func (mr *MockSegmentUsersCacheMockRecorder) GetSegmentIDs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSegmentIDs", reflect.TypeOf((*MockSegmentUsersCache)(nil).GetSegmentIDs))
}

// Put mocks base method.
func (m *MockSegmentUsersCache) Put(segmentUsers *feature.SegmentUsers) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", segmentUsers)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockSegmentUsersCacheMockRecorder) Put(segmentUsers interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockSegmentUsersCache)(nil).Put), segmentUsers)
}
