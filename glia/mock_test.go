// Code generated by MockGen. DO NOT EDIT.
// Source: glia.go

// Package glia_test is a generated GoMock package.
package glia_test

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	glia "github.com/wetware/go/glia"
	proc "github.com/wetware/go/proc"
)

// MockProc is a mock of Proc interface.
type MockProc struct {
	ctrl     *gomock.Controller
	recorder *MockProcMockRecorder
}

// MockProcMockRecorder is the mock recorder for MockProc.
type MockProcMockRecorder struct {
	mock *MockProc
}

// NewMockProc creates a new mock instance.
func NewMockProc(ctrl *gomock.Controller) *MockProc {
	mock := &MockProc{ctrl: ctrl}
	mock.recorder = &MockProcMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProc) EXPECT() *MockProcMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockProc) Close(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockProcMockRecorder) Close(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockProc)(nil).Close), arg0)
}

// Method mocks base method.
func (m *MockProc) Method(name string) proc.Method {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Method", name)
	ret0, _ := ret[0].(proc.Method)
	return ret0
}

// Method indicates an expected call of Method.
func (mr *MockProcMockRecorder) Method(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Method", reflect.TypeOf((*MockProc)(nil).Method), name)
}

// Release mocks base method.
func (m *MockProc) Release() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Release")
}

// Release indicates an expected call of Release.
func (mr *MockProcMockRecorder) Release() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Release", reflect.TypeOf((*MockProc)(nil).Release))
}

// Reserve mocks base method.
func (m *MockProc) Reserve(ctx context.Context, body io.Reader) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reserve", ctx, body)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reserve indicates an expected call of Reserve.
func (mr *MockProcMockRecorder) Reserve(ctx, body interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reserve", reflect.TypeOf((*MockProc)(nil).Reserve), ctx, body)
}

// String mocks base method.
func (m *MockProc) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockProcMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockProc)(nil).String))
}

// MockRouter is a mock of Router interface.
type MockRouter struct {
	ctrl     *gomock.Controller
	recorder *MockRouterMockRecorder
}

// MockRouterMockRecorder is the mock recorder for MockRouter.
type MockRouterMockRecorder struct {
	mock *MockRouter
}

// NewMockRouter creates a new mock instance.
func NewMockRouter(ctrl *gomock.Controller) *MockRouter {
	mock := &MockRouter{ctrl: ctrl}
	mock.recorder = &MockRouterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRouter) EXPECT() *MockRouterMockRecorder {
	return m.recorder
}

// GetProc mocks base method.
func (m *MockRouter) GetProc(pid string) (glia.Proc, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProc", pid)
	ret0, _ := ret[0].(glia.Proc)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProc indicates an expected call of GetProc.
func (mr *MockRouterMockRecorder) GetProc(pid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProc", reflect.TypeOf((*MockRouter)(nil).GetProc), pid)
}