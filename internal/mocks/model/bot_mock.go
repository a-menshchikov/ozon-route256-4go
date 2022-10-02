// Code generated by MockGen. DO NOT EDIT.
// Source: internal/model/bot.go

// Package mock_model is a generated GoMock package.
package mock_model

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockMessageSender is a mock of MessageSender interface.
type MockMessageSender struct {
	ctrl     *gomock.Controller
	recorder *MockMessageSenderMockRecorder
}

// MockMessageSenderMockRecorder is the mock recorder for MockMessageSender.
type MockMessageSenderMockRecorder struct {
	mock *MockMessageSender
}

// NewMockMessageSender creates a new mock instance.
func NewMockMessageSender(ctrl *gomock.Controller) *MockMessageSender {
	mock := &MockMessageSender{ctrl: ctrl}
	mock.recorder = &MockMessageSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessageSender) EXPECT() *MockMessageSenderMockRecorder {
	return m.recorder
}

// SendMessage mocks base method.
func (m *MockMessageSender) SendMessage(userID int64, text string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessage", userID, text)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMessage indicates an expected call of SendMessage.
func (mr *MockMessageSenderMockRecorder) SendMessage(userID, text interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessage", reflect.TypeOf((*MockMessageSender)(nil).SendMessage), userID, text)
}

// MockExpenseStorage is a mock of ExpenseStorage interface.
type MockExpenseStorage struct {
	ctrl     *gomock.Controller
	recorder *MockExpenseStorageMockRecorder
}

// MockExpenseStorageMockRecorder is the mock recorder for MockExpenseStorage.
type MockExpenseStorageMockRecorder struct {
	mock *MockExpenseStorage
}

// NewMockExpenseStorage creates a new mock instance.
func NewMockExpenseStorage(ctrl *gomock.Controller) *MockExpenseStorage {
	mock := &MockExpenseStorage{ctrl: ctrl}
	mock.recorder = &MockExpenseStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExpenseStorage) EXPECT() *MockExpenseStorageMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockExpenseStorage) Add(userID int64, date time.Time, amount int64, category string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", userID, date, amount, category)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockExpenseStorageMockRecorder) Add(userID, date, amount, category interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockExpenseStorage)(nil).Add), userID, date, amount, category)
}

// Init mocks base method.
func (m *MockExpenseStorage) Init(userID int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Init", userID)
}

// Init indicates an expected call of Init.
func (mr *MockExpenseStorageMockRecorder) Init(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockExpenseStorage)(nil).Init), userID)
}

// List mocks base method.
func (m *MockExpenseStorage) List(userID int64, from time.Time) map[string]int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", userID, from)
	ret0, _ := ret[0].(map[string]int64)
	return ret0
}

// List indicates an expected call of List.
func (mr *MockExpenseStorageMockRecorder) List(userID, from interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockExpenseStorage)(nil).List), userID, from)
}
