package model

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
	today       = utils.TruncateToDate(time.Now())
	yesterday   = today.Add(-time.Hour * 24)
	simpleError = errors.New("error")
)

type mocksInitializer struct {
	expenser        func(*mocks.Mockexpenser)
	limiter         func(*mocks.Mocklimiter)
	currencyManager func(*mocks.MockcurrencyManager)
	rater           func(*mocks.Mockrater)
}

func setupController(t *testing.T, i mocksInitializer) *controller {
	ctrl := gomock.NewController(t)

	expenserMock := mocks.NewMockexpenser(ctrl)
	if i.expenser != nil {
		i.expenser(expenserMock)
	}

	limiterMock := mocks.NewMocklimiter(ctrl)
	if i.limiter != nil {
		i.limiter(limiterMock)
	}

	currencyManagerMock := mocks.NewMockcurrencyManager(ctrl)
	if i.currencyManager != nil {
		i.currencyManager(currencyManagerMock)
	}

	raterMock := mocks.NewMockrater(ctrl)
	if i.rater != nil {
		i.rater(raterMock)
	}

	return NewController(expenserMock, limiterMock, currencyManagerMock, raterMock, zap.NewNop())
}

func Test_controller_ListCurrencies(t *testing.T) {
	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("", simpleError)
			},
		})

		// ACT
		resp := controller.ListCurrencies(request.ListCurrencies{
			User: testUser,
		})

		// ASSERT
		assert.Empty(t, resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
				m.EXPECT().ListCurrenciesCodesWithFlags().Return([]string{"RUB", "USD"})
			},
		})

		// ACT
		resp := controller.ListCurrencies(request.ListCurrencies{
			User: testUser,
		})

		// ASSERT
		assert.Equal(t, response.ListCurrencies{
			Current: "RUB",
			List:    []string{"RUB", "USD"},
		}, resp)
	})
}

func Test_controller_SetCurrency(t *testing.T) {
	t.Run("failed", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Set(testUser, "RUB").Return(simpleError)
			},
		})

		// ACT
		resp := controller.SetCurrency(request.SetCurrency{
			User: testUser,
			Code: "RUB",
		})

		// ASSERT
		assert.Equal(t, response.SetCurrency(false), resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Set(testUser, "EUR").Return(nil)
			},
		})

		// ACT
		resp := controller.SetCurrency(request.SetCurrency{
			User: testUser,
			Code: "EUR",
		})

		// ASSERT
		assert.Equal(t, response.SetCurrency(true), resp)
	})
}

func Test_controller_ListLimits(t *testing.T) {
	t.Run("not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(false)
			},
		})

		// ACT
		resp := controller.ListLimits(request.ListLimits{
			User: testUser,
		})

		// ASSERT
		assert.Empty(t, resp)
	})

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("", simpleError)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.ListLimits(request.ListLimits{
			User: testUser,
		})

		// ASSERT
		assert.Equal(t, response.ListLimits{
			Ready: true,
		}, resp)
	})

	t.Run("no limits list", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(testUser).Return(nil, simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.ListLimits(request.ListLimits{
			User: testUser,
		})

		// ASSERT
		assert.Equal(t, response.ListLimits{
			Ready:           true,
			CurrentCurrency: "RUB",
		}, resp)
	})

	t.Run("cannot exchange total", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
					"": {
						Total:    1000000,
						Remains:  750000,
						Currency: "USD",
					},
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(1000000), "USD", "RUB", today).Return(int64(0), simpleError)
			},
		})

		// ACT
		resp := controller.ListLimits(request.ListLimits{
			User: testUser,
		})

		// ASSERT
		assert.Equal(t, response.ListLimits{
			Ready:           true,
			CurrentCurrency: "RUB",
		}, resp)
	})

	t.Run("cannot exchange remains", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
					"": {
						Total:    2000000,
						Remains:  1500000,
						Currency: "USD",
					},
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(2000000), "USD", "RUB", today).Return(int64(10000000), nil)
				m.EXPECT().Exchange(int64(1500000), "USD", "RUB", today).Return(int64(0), simpleError)
			},
		})

		// ACT
		resp := controller.ListLimits(request.ListLimits{
			User: testUser,
		})

		// ASSERT
		assert.Equal(t, response.ListLimits{
			Ready:           true,
			CurrentCurrency: "RUB",
		}, resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
					"taxi": {
						Total:    400000,
						Remains:  300000,
						Currency: "USD",
					},
					"": {
						Total:    30000000,
						Remains:  10000000,
						Currency: "RUB",
					},
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(400000), "USD", "RUB", today).Return(int64(2000000), nil)
				m.EXPECT().Exchange(int64(300000), "USD", "RUB", today).Return(int64(1500000), nil)
				m.EXPECT().Exchange(int64(30000000), "RUB", "RUB", today).Return(int64(30000000), nil)
				m.EXPECT().Exchange(int64(10000000), "RUB", "RUB", today).Return(int64(10000000), nil)
			},
		})

		// ACT
		resp := controller.ListLimits(request.ListLimits{
			User: testUser,
		})

		// ASSERT
		assert.Equal(t, response.ListLimits{
			Ready:           true,
			CurrentCurrency: "RUB",
			List: map[string]response.LimitItem{
				"taxi": {
					Total:   2000000,
					Remains: 1500000,
					Origin: types.LimitItem{
						Total:    400000,
						Remains:  300000,
						Currency: "USD",
					},
				},
				"": {
					Total:   30000000,
					Remains: 10000000,
					Origin: types.LimitItem{
						Total:    30000000,
						Remains:  10000000,
						Currency: "RUB",
					},
				},
			},
			Success: true,
		}, resp)
	})
}

func Test_controller_SetLimit(t *testing.T) {
	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("", simpleError)
			},
		})

		// ACT
		resp := controller.SetLimit(request.SetLimit{
			User:     testUser,
			Value:    1000000,
			Category: "",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(false), resp)
	})

	t.Run("cannot set", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Set(testUser, int64(1000000), "USD", "taxi").Return(simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(request.SetLimit{
			User:     testUser,
			Value:    1000000,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(false), resp)
	})

	t.Run("successful set", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Set(testUser, int64(2000000), "USD", "taxi").Return(nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(request.SetLimit{
			User:     testUser,
			Value:    2000000,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(true), resp)
	})

	t.Run("cannot unset", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Unset(testUser, "taxi").Return(simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(request.SetLimit{
			User:     testUser,
			Value:    0,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(false), resp)
	})

	t.Run("successful unset", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Unset(testUser, "taxi").Return(nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(request.SetLimit{
			User:     testUser,
			Value:    0,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(true), resp)
	})
}

func Test_controller_AddExpense(t *testing.T) {
	t.Run("not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(false)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   10000,
			Category: "coffee",
		})

		// ASSERT
		assert.Empty(t, resp)
	})

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("", simpleError)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   15000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready: true,
		}, resp)
	})

	t.Run("cannot add", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(20000), "USD", "coffee").Return(simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   20000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready: true,
		}, resp)
	})

	t.Run("cannot get limits", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(25000), "USD", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{}, simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   25000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready:   true,
			Success: true,
		}, resp)
	})

	t.Run("no limit", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(30000), "USD", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
					Total: 0,
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   30000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready:   true,
			Success: true,
		}, resp)
	})

	t.Run("cannot exchange for limit", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(35000), "EUR", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("EUR", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(35000), "EUR", "USD", today).Return(int64(0), simpleError)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   35000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready:   true,
			Success: true,
		}, resp)
	})

	t.Run("cannot decrease limit", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(2000000), "RUB", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
				m.EXPECT().Decrease(testUser, int64(40000), "coffee").Return(false, simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(2000000), "RUB", "USD", today).Return(int64(40000), nil)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   2000000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready:   true,
			Success: true,
		}, resp)
	})

	t.Run("limit has been reached", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(2000000), "RUB", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
				m.EXPECT().Decrease(testUser, int64(40000), "coffee").Return(true, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(2000000), "RUB", "USD", today).Return(int64(40000), nil)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   2000000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready:        true,
			LimitReached: true,
			Success:      true,
		}, resp)
	})

	t.Run("limit has not been reached", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Add(testUser, today, int64(1000000), "RUB", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
				m.EXPECT().Decrease(testUser, int64(20000), "coffee").Return(false, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(1000000), "RUB", "USD", today).Return(int64(20000), nil)
			},
		})

		// ACT
		resp := controller.AddExpense(request.AddExpense{
			User:     testUser,
			Date:     today,
			Amount:   1000000,
			Category: "coffee",
		})

		// ASSERT
		assert.Equal(t, response.AddExpense{
			Ready:        true,
			LimitReached: false,
			Success:      true,
		}, resp)
	})
}

func Test_controller_GetReport(t *testing.T) {
	t.Run("not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(false)
			},
		})

		// ACT
		resp := controller.GetReport(request.GetReport{
			User: testUser,
			From: today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:  today,
			Ready: false,
		}, resp)
	})

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("", simpleError)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.GetReport(request.GetReport{
			User: testUser,
			From: today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:  today,
			Ready: true,
		}, resp)
	})

	t.Run("no data", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Report(testUser, today).Return(nil, simpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("RUB", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
			},
		})

		// ACT
		resp := controller.GetReport(request.GetReport{
			User: testUser,
			From: today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:     today,
			Ready:    true,
			Currency: "RUB",
		}, resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, mocksInitializer{
			expenser: func(m *mocks.Mockexpenser) {
				m.EXPECT().Report(testUser, today).Return(map[string][]types.ExpenseItem{
					"coffee": {
						{
							Date:     today,
							Amount:   1000000,
							Currency: "RUB",
						},
					},
					"taxi": {
						{
							Date:     today,
							Amount:   100000,
							Currency: "USD",
						},
						{
							Date:     yesterday,
							Amount:   1500000,
							Currency: "RUB",
						},
					},
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(testUser).Return("USD", nil)
			},
			rater: func(m *mocks.Mockrater) {
				m.EXPECT().Ready().Return(true)
				m.EXPECT().Exchange(int64(1000000), "RUB", "USD", today).Return(int64(20000), nil)
				m.EXPECT().Exchange(int64(100000), "USD", "USD", today).Return(int64(100000), nil)
				m.EXPECT().Exchange(int64(1500000), "RUB", "USD", yesterday).Return(int64(30000), nil)
			},
		})

		// ACT
		resp := controller.GetReport(request.GetReport{
			User: testUser,
			From: today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:     today,
			Ready:    true,
			Currency: "USD",
			Data: map[string]int64{
				"coffee": 20000,
				"taxi":   130000,
			},
			Success: true,
		}, resp)
	})
}
