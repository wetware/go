// Code generated by MockGen. DO NOT EDIT.
// Source: glia.go
//
// Generated by this command:
//
//	mockgen -source=glia.go -destination=mock_test.go -package=glia_test
//

// Package glia_test is a generated GoMock package.
package glia_test

import (
	context "context"
	io "io"
	slog "log/slog"
	reflect "reflect"

	host "github.com/libp2p/go-libp2p/core/host"
	protocol "github.com/libp2p/go-libp2p/core/protocol"
	routing "github.com/libp2p/go-libp2p/core/routing"
	proc "github.com/wetware/go/proc"
	gomock "go.uber.org/mock/gomock"
)

// MockEnv is a mock of Env interface.
type MockEnv struct {
	ctrl     *gomock.Controller
	recorder *MockEnvMockRecorder
	isgomock struct{}
}

// MockEnvMockRecorder is the mock recorder for MockEnv.
type MockEnvMockRecorder struct {
	mock *MockEnv
}

// NewMockEnv creates a new mock instance.
func NewMockEnv(ctrl *gomock.Controller) *MockEnv {
	mock := &MockEnv{ctrl: ctrl}
	mock.recorder = &MockEnvMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEnv) EXPECT() *MockEnvMockRecorder {
	return m.recorder
}

// LocalHost mocks base method.
func (m *MockEnv) LocalHost() host.Host {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LocalHost")
	ret0, _ := ret[0].(host.Host)
	return ret0
}

// LocalHost indicates an expected call of LocalHost.
func (mr *MockEnvMockRecorder) LocalHost() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalHost", reflect.TypeOf((*MockEnv)(nil).LocalHost))
}

// Log mocks base method.
func (m *MockEnv) Log() *slog.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Log")
	ret0, _ := ret[0].(*slog.Logger)
	return ret0
}

// Log indicates an expected call of Log.
func (mr *MockEnvMockRecorder) Log() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockEnv)(nil).Log))
}

// Routing mocks base method.
func (m *MockEnv) Routing() routing.Routing {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Routing")
	ret0, _ := ret[0].(routing.Routing)
	return ret0
}

// Routing indicates an expected call of Routing.
func (mr *MockEnvMockRecorder) Routing() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Routing", reflect.TypeOf((*MockEnv)(nil).Routing))
}

// MockProc is a mock of Proc interface.
type MockProc struct {
	ctrl     *gomock.Controller
	recorder *MockProcMockRecorder
	isgomock struct{}
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
func (mr *MockProcMockRecorder) Close(arg0 any) *gomock.Call {
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
func (mr *MockProcMockRecorder) Method(name any) *gomock.Call {
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
func (m *MockProc) Reserve(arg0 context.Context, arg1 io.ReadWriteCloser) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Reserve", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reserve indicates an expected call of Reserve.
func (mr *MockProcMockRecorder) Reserve(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reserve", reflect.TypeOf((*MockProc)(nil).Reserve), arg0, arg1)
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

// MockStream is a mock of Stream interface.
type MockStream struct {
	ctrl     *gomock.Controller
	recorder *MockStreamMockRecorder
	isgomock struct{}
}

// MockStreamMockRecorder is the mock recorder for MockStream.
type MockStreamMockRecorder struct {
	mock *MockStream
}

// NewMockStream creates a new mock instance.
func NewMockStream(ctrl *gomock.Controller) *MockStream {
	mock := &MockStream{ctrl: ctrl}
	mock.recorder = &MockStreamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStream) EXPECT() *MockStreamMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockStream) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStreamMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStream)(nil).Close))
}

// CloseRead mocks base method.
func (m *MockStream) CloseRead() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseRead")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseRead indicates an expected call of CloseRead.
func (mr *MockStreamMockRecorder) CloseRead() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseRead", reflect.TypeOf((*MockStream)(nil).CloseRead))
}

// CloseWrite mocks base method.
func (m *MockStream) CloseWrite() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseWrite")
	ret0, _ := ret[0].(error)
	return ret0
}

// CloseWrite indicates an expected call of CloseWrite.
func (mr *MockStreamMockRecorder) CloseWrite() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseWrite", reflect.TypeOf((*MockStream)(nil).CloseWrite))
}

// Destination mocks base method.
func (m *MockStream) Destination() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Destination")
	ret0, _ := ret[0].(string)
	return ret0
}

// Destination indicates an expected call of Destination.
func (mr *MockStreamMockRecorder) Destination() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destination", reflect.TypeOf((*MockStream)(nil).Destination))
}

// MethodName mocks base method.
func (m *MockStream) MethodName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MethodName")
	ret0, _ := ret[0].(string)
	return ret0
}

// MethodName indicates an expected call of MethodName.
func (mr *MockStreamMockRecorder) MethodName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MethodName", reflect.TypeOf((*MockStream)(nil).MethodName))
}

// ProcID mocks base method.
func (m *MockStream) ProcID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ProcID indicates an expected call of ProcID.
func (mr *MockStreamMockRecorder) ProcID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcID", reflect.TypeOf((*MockStream)(nil).ProcID))
}

// Protocol mocks base method.
func (m *MockStream) Protocol() protocol.ID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Protocol")
	ret0, _ := ret[0].(protocol.ID)
	return ret0
}

// Protocol indicates an expected call of Protocol.
func (mr *MockStreamMockRecorder) Protocol() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Protocol", reflect.TypeOf((*MockStream)(nil).Protocol))
}

// Read mocks base method.
func (m *MockStream) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockStreamMockRecorder) Read(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockStream)(nil).Read), p)
}

// Write mocks base method.
func (m *MockStream) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockStreamMockRecorder) Write(p any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockStream)(nil).Write), p)
}
