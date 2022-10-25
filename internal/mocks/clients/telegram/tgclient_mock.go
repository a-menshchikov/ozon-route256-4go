// Code generated by MockGen. DO NOT EDIT.
// Source: internal/clients/telegram/tgclient.go

// Package mock_telegram is a generated GoMock package.
package mock_telegram

import (
	reflect "reflect"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gomock "github.com/golang/mock/gomock"
)

// Mockapi is a mock of api interface.
type Mockapi struct {
	ctrl     *gomock.Controller
	recorder *MockapiMockRecorder
}

// MockapiMockRecorder is the mock recorder for Mockapi.
type MockapiMockRecorder struct {
	mock *Mockapi
}

// NewMockapi creates a new mock instance.
func NewMockapi(ctrl *gomock.Controller) *Mockapi {
	mock := &Mockapi{ctrl: ctrl}
	mock.recorder = &MockapiMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockapi) EXPECT() *MockapiMockRecorder {
	return m.recorder
}

// GetUpdatesChan mocks base method.
func (m *Mockapi) GetUpdatesChan(arg0 tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdatesChan", arg0)
	ret0, _ := ret[0].(tgbotapi.UpdatesChannel)
	return ret0
}

// GetUpdatesChan indicates an expected call of GetUpdatesChan.
func (mr *MockapiMockRecorder) GetUpdatesChan(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdatesChan", reflect.TypeOf((*Mockapi)(nil).GetUpdatesChan), arg0)
}

// Request mocks base method.
func (m *Mockapi) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Request", c)
	ret0, _ := ret[0].(*tgbotapi.APIResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Request indicates an expected call of Request.
func (mr *MockapiMockRecorder) Request(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Request", reflect.TypeOf((*Mockapi)(nil).Request), c)
}

// Send mocks base method.
func (m *Mockapi) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", c)
	ret0, _ := ret[0].(tgbotapi.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send.
func (mr *MockapiMockRecorder) Send(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*Mockapi)(nil).Send), c)
}
