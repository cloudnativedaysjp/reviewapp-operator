// Code generated by MockGen. DO NOT EDIT.
// Source: gateways/git.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gateways "github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	gomock "github.com/golang/mock/gomock"
)

// MockGitIFace is a mock of GitIFace interface.
type MockGitIFace struct {
	ctrl     *gomock.Controller
	recorder *MockGitIFaceMockRecorder
}

// MockGitIFaceMockRecorder is the mock recorder for MockGitIFace.
type MockGitIFaceMockRecorder struct {
	mock *MockGitIFace
}

// NewMockGitIFace creates a new mock instance.
func NewMockGitIFace(ctrl *gomock.Controller) *MockGitIFace {
	mock := &MockGitIFace{ctrl: ctrl}
	mock.recorder = &MockGitIFaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitIFace) EXPECT() *MockGitIFaceMockRecorder {
	return m.recorder
}

// WithCredential mocks base method.
func (m *MockGitIFace) WithCredential(username, token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithCredential", username, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// WithCredential indicates an expected call of WithCredential.
func (mr *MockGitIFaceMockRecorder) WithCredential(username, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithCredential", reflect.TypeOf((*MockGitIFace)(nil).WithCredential), username, token)
}

// ForceClone mocks base method.
func (m *MockGitIFace) ForceClone(ctx context.Context, org, repo, branch string) (*gateways.GitProject, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceClone", ctx, org, repo, branch)
	ret0, _ := ret[0].(*gateways.GitProject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ForceClone indicates an expected call of ForceClone.
func (mr *MockGitIFaceMockRecorder) ForceClone(ctx, org, repo, branch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceClone", reflect.TypeOf((*MockGitIFace)(nil).ForceClone), ctx, org, repo, branch)
}

// CreateFile mocks base method.
func (m *MockGitIFace) CreateFile(ctx context.Context, gp gateways.GitProject, filename string, contents []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFile", ctx, gp, filename, contents)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateFile indicates an expected call of CreateFile.
func (mr *MockGitIFaceMockRecorder) CreateFile(ctx, gp, filename, contents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFile", reflect.TypeOf((*MockGitIFace)(nil).CreateFile), ctx, gp, filename, contents)
}

// DeleteFile mocks base method.
func (m *MockGitIFace) DeleteFile(ctx context.Context, gp gateways.GitProject, filename string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFile", ctx, gp, filename)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFile indicates an expected call of DeleteFile.
func (mr *MockGitIFaceMockRecorder) DeleteFile(ctx, gp, filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFile", reflect.TypeOf((*MockGitIFace)(nil).DeleteFile), ctx, gp, filename)
}

// CommitAndPush mocks base method.
func (m *MockGitIFace) CommitAndPush(ctx context.Context, gp gateways.GitProject, message string) (*gateways.GitProject, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitAndPush", ctx, gp, message)
	ret0, _ := ret[0].(*gateways.GitProject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CommitAndPush indicates an expected call of CommitAndPush.
func (mr *MockGitIFaceMockRecorder) CommitAndPush(ctx, gp, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitAndPush", reflect.TypeOf((*MockGitIFace)(nil).CommitAndPush), ctx, gp, message)
}
