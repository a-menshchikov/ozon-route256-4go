// Code generated by MockGen. DO NOT EDIT.
// Source: internal/model/currency/rater.go

// Package mock_currency is a generated GoMock package.
package mock_currency

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// Mockgateway is a mock of gateway interface.
type Mockgateway struct {
	ctrl     *gomock.Controller
	recorder *MockgatewayMockRecorder
}

// MockgatewayMockRecorder is the mock recorder for Mockgateway.
type MockgatewayMockRecorder struct {
	mock *Mockgateway
}

// NewMockgateway creates a new mock instance.
func NewMockgateway(ctrl *gomock.Controller) *Mockgateway {
	mock := &Mockgateway{ctrl: ctrl}
	mock.recorder = &MockgatewayMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockgateway) EXPECT() *MockgatewayMockRecorder {
	return m.recorder
}

// FetchRates mocks base method.
func (m *Mockgateway) FetchRates(ctx context.Context) (map[string]int64, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchRates", ctx)
	ret0, _ := ret[0].(map[string]int64)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// FetchRates indicates an expected call of FetchRates.
func (mr *MockgatewayMockRecorder) FetchRates(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchRates", reflect.TypeOf((*Mockgateway)(nil).FetchRates), ctx)
}
