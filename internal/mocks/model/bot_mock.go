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

// SendMessageWithInlineKeyboard mocks base method.
func (m *MockMessageSender) SendMessageWithInlineKeyboard(userID int64, text string, rows [][][]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessageWithInlineKeyboard", userID, text, rows)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMessageWithInlineKeyboard indicates an expected call of SendMessageWithInlineKeyboard.
func (mr *MockMessageSenderMockRecorder) SendMessageWithInlineKeyboard(userID, text, rows interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessageWithInlineKeyboard", reflect.TypeOf((*MockMessageSender)(nil).SendMessageWithInlineKeyboard), userID, text, rows)
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

// MockExchanger is a mock of Exchanger interface.
type MockExchanger struct {
	ctrl     *gomock.Controller
	recorder *MockExchangerMockRecorder
}

// MockExchangerMockRecorder is the mock recorder for MockExchanger.
type MockExchangerMockRecorder struct {
	mock *MockExchanger
}

// NewMockExchanger creates a new mock instance.
func NewMockExchanger(ctrl *gomock.Controller) *MockExchanger {
	mock := &MockExchanger{ctrl: ctrl}
	mock.recorder = &MockExchangerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExchanger) EXPECT() *MockExchangerMockRecorder {
	return m.recorder
}

// ExchangeFromBase mocks base method.
func (m *MockExchanger) ExchangeFromBase(value int64, currency string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExchangeFromBase", value, currency)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExchangeFromBase indicates an expected call of ExchangeFromBase.
func (mr *MockExchangerMockRecorder) ExchangeFromBase(value, currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExchangeFromBase", reflect.TypeOf((*MockExchanger)(nil).ExchangeFromBase), value, currency)
}

// ExchangeToBase mocks base method.
func (m *MockExchanger) ExchangeToBase(value int64, currency string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExchangeToBase", value, currency)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExchangeToBase indicates an expected call of ExchangeToBase.
func (mr *MockExchangerMockRecorder) ExchangeToBase(value, currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExchangeToBase", reflect.TypeOf((*MockExchanger)(nil).ExchangeToBase), value, currency)
}

// ListCurrencies mocks base method.
func (m *MockExchanger) ListCurrencies() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCurrencies")
	ret0, _ := ret[0].([]string)
	return ret0
}

// ListCurrencies indicates an expected call of ListCurrencies.
func (mr *MockExchangerMockRecorder) ListCurrencies() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListCurrencies", reflect.TypeOf((*MockExchanger)(nil).ListCurrencies))
}

// Ready mocks base method.
func (m *MockExchanger) Ready() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ready")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Ready indicates an expected call of Ready.
func (mr *MockExchangerMockRecorder) Ready() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ready", reflect.TypeOf((*MockExchanger)(nil).Ready))
}

// MockCurrencyKeeper is a mock of CurrencyKeeper interface.
type MockCurrencyKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockCurrencyKeeperMockRecorder
}

// MockCurrencyKeeperMockRecorder is the mock recorder for MockCurrencyKeeper.
type MockCurrencyKeeperMockRecorder struct {
	mock *MockCurrencyKeeper
}

// NewMockCurrencyKeeper creates a new mock instance.
func NewMockCurrencyKeeper(ctrl *gomock.Controller) *MockCurrencyKeeper {
	mock := &MockCurrencyKeeper{ctrl: ctrl}
	mock.recorder = &MockCurrencyKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCurrencyKeeper) EXPECT() *MockCurrencyKeeperMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockCurrencyKeeper) Get(userID int64) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", userID)
	ret0, _ := ret[0].(string)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockCurrencyKeeperMockRecorder) Get(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCurrencyKeeper)(nil).Get), userID)
}

// Set mocks base method.
func (m *MockCurrencyKeeper) Set(userID int64, currency string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", userID, currency)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCurrencyKeeperMockRecorder) Set(userID, currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCurrencyKeeper)(nil).Set), userID, currency)
}