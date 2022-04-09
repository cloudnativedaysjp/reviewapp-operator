// Code generated by MockGen. DO NOT EDIT.
// Source: ./domain/repositories/gitcommand.go

// Package mock_repositories is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	models "github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	gomock "github.com/golang/mock/gomock"
)

// MockGitCommand is a mock of GitCommand interface.
type MockGitCommand struct {
	ctrl     *gomock.Controller
	recorder *MockGitCommandMockRecorder
}

// MockGitCommandMockRecorder is the mock recorder for MockGitCommand.
type MockGitCommandMockRecorder struct {
	mock *MockGitCommand
}

// NewMockGitCommand creates a new mock instance.
func NewMockGitCommand(ctrl *gomock.Controller) *MockGitCommand {
	mock := &MockGitCommand{ctrl: ctrl}
	mock.recorder = &MockGitCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGitCommand) EXPECT() *MockGitCommandMockRecorder {
	return m.recorder
}

// CommitAndPush mocks base method.
func (m *MockGitCommand) CommitAndPush(ctx context.Context, gp models.InfraRepoLocalDir, message string) (*models.InfraRepoLocalDir, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitAndPush", ctx, gp, message)
	ret0, _ := ret[0].(*models.InfraRepoLocalDir)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CommitAndPush indicates an expected call of CommitAndPush.
func (mr *MockGitCommandMockRecorder) CommitAndPush(ctx, gp, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitAndPush", reflect.TypeOf((*MockGitCommand)(nil).CommitAndPush), ctx, gp, message)
}

// CreateFiles mocks base method.
func (m *MockGitCommand) CreateFiles(arg0 context.Context, arg1 models.InfraRepoLocalDir, arg2 ...models.File) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateFiles", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateFiles indicates an expected call of CreateFiles.
func (mr *MockGitCommandMockRecorder) CreateFiles(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFiles", reflect.TypeOf((*MockGitCommand)(nil).CreateFiles), varargs...)
}

// DeleteFiles mocks base method.
func (m *MockGitCommand) DeleteFiles(arg0 context.Context, arg1 models.InfraRepoLocalDir, arg2 ...models.File) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteFiles", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFiles indicates an expected call of DeleteFiles.
func (mr *MockGitCommandMockRecorder) DeleteFiles(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFiles", reflect.TypeOf((*MockGitCommand)(nil).DeleteFiles), varargs...)
}

// ForceClone mocks base method.
func (m *MockGitCommand) ForceClone(arg0 context.Context, arg1 models.InfraRepoTarget) (models.InfraRepoLocalDir, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ForceClone", arg0, arg1)
	ret0, _ := ret[0].(models.InfraRepoLocalDir)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ForceClone indicates an expected call of ForceClone.
func (mr *MockGitCommandMockRecorder) ForceClone(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ForceClone", reflect.TypeOf((*MockGitCommand)(nil).ForceClone), arg0, arg1)
}

// WithCredential mocks base method.
func (m *MockGitCommand) WithCredential(credential models.GitCredential) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithCredential", credential)
	ret0, _ := ret[0].(error)
	return ret0
}

// WithCredential indicates an expected call of WithCredential.
func (mr *MockGitCommandMockRecorder) WithCredential(credential interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithCredential", reflect.TypeOf((*MockGitCommand)(nil).WithCredential), credential)
}