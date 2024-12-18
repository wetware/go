// Code generated by MockGen. DO NOT EDIT.
// Source: proc.go

// Package proc_test is a generated GoMock package.
package proc_test

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMethod is a mock of Method interface.
type MockMethod struct {
	ctrl     *gomock.Controller
	recorder *MockMethodMockRecorder
}

// MockMethodMockRecorder is the mock recorder for MockMethod.
type MockMethodMockRecorder struct {
	mock *MockMethod
}

// NewMockMethod creates a new mock instance.
func NewMockMethod(ctrl *gomock.Controller) *MockMethod {
	mock := &MockMethod{ctrl: ctrl}
	mock.recorder = &MockMethodMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMethod) EXPECT() *MockMethodMockRecorder {
	return m.recorder
}

// CallWithStack mocks base method.
func (m *MockMethod) CallWithStack(arg0 context.Context, arg1 []uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CallWithStack", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CallWithStack indicates an expected call of CallWithStack.
func (mr *MockMethodMockRecorder) CallWithStack(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CallWithStack", reflect.TypeOf((*MockMethod)(nil).CallWithStack), arg0, arg1)
}