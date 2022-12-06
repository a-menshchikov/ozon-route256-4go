//go:build unit

package model

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type controllerMocksInitializer struct {
	expenser        func(m *mocks.MockExpenser)
	reporter        func(m *mocks.MockReporter)
	limiter         func(m *mocks.Mocklimiter)
	currencyManager func(m *mocks.MockcurrencyManager)
	rater           func(m *mocks.MockRater)
}

func setupController(t *testing.T, i controllerMocksInitializer) *controller {
	ctrl := gomock.NewController(t)

	expenserMock := mocks.NewMockExpenser(ctrl)
	if i.expenser != nil {
		i.expenser(expenserMock)
	}

	reporterMock := mocks.NewMockReporter(ctrl)
	if i.reporter != nil {
		i.reporter(reporterMock)
	}

	limiterMock := mocks.NewMocklimiter(ctrl)
	if i.limiter != nil {
		i.limiter(limiterMock)
	}

	currencyManagerMock := mocks.NewMockcurrencyManager(ctrl)
	if i.currencyManager != nil {
		i.currencyManager(currencyManagerMock)
	}

	raterMock := mocks.NewMockRater(ctrl)
	if i.rater != nil {
		i.rater(raterMock)
	}

	return NewController(expenserMock, reporterMock, limiterMock, currencyManagerMock, raterMock, zap.NewNop())
}

func Test_controller_ListCurrencies(t *testing.T) {
	t.Parallel()

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", test.SimpleError)
			},
		})

		// ACT
		resp := controller.ListCurrencies(context.Background(), request.ListCurrencies{
			User: test.User,
		})

		// ASSERT
		assert.Empty(t, resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
				m.EXPECT().ListCurrenciesCodesWithFlags().Return([]string{"RUB", "USD"})
			},
		})

		// ACT
		resp := controller.ListCurrencies(context.Background(), request.ListCurrencies{
			User: test.User,
		})

		// ASSERT
		assert.Equal(t, response.ListCurrencies{
			Current: "RUB",
			List:    []string{"RUB", "USD"},
		}, resp)
	})
}

func Test_controller_SetCurrency(t *testing.T) {
	t.Parallel()

	t.Run("failed", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "RUB").Return(test.SimpleError)
			},
		})

		// ACT
		resp := controller.SetCurrency(context.Background(), request.SetCurrency{
			User: test.User,
			Code: "RUB",
		})

		// ASSERT
		assert.Equal(t, response.SetCurrency(false), resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "EUR").Return(nil)
			},
		})

		// ACT
		resp := controller.SetCurrency(context.Background(), request.SetCurrency{
			User: test.User,
			Code: "EUR",
		})

		// ASSERT
		assert.Equal(t, response.SetCurrency(true), resp)
	})
}

func Test_controller_ListLimits(t *testing.T) {
	t.Parallel()

	t.Run("not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(false)
			},
		})

		// ACT
		resp := controller.ListLimits(context.Background(), request.ListLimits{
			User: test.User,
		})

		// ASSERT
		assert.Empty(t, resp)
	})

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", test.SimpleError)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		resp := controller.ListLimits(context.Background(), request.ListLimits{
			User: test.User,
		})

		// ASSERT
		assert.Equal(t, response.ListLimits{
			Ready: true,
		}, resp)
	})

	t.Run("no limits list", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return(nil, test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		resp := controller.ListLimits(context.Background(), request.ListLimits{
			User: test.User,
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
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return(map[string]types.LimitItem{
					"": {
						Total:    1000000,
						Remains:  750000,
						Currency: "USD",
					},
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(1000000), "USD", "RUB", test.Today).Return(int64(0), test.SimpleError)
			},
		})

		// ACT
		resp := controller.ListLimits(context.Background(), request.ListLimits{
			User: test.User,
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
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return(map[string]types.LimitItem{
					"": {
						Total:    2000000,
						Remains:  1500000,
						Currency: "USD",
					},
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(2000000), "USD", "RUB", test.Today).Return(int64(1000000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(1500000), "USD", "RUB", test.Today).Return(int64(0), test.SimpleError)
			},
		})

		// ACT
		resp := controller.ListLimits(context.Background(), request.ListLimits{
			User: test.User,
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
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return(map[string]types.LimitItem{
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
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(400000), "USD", "RUB", test.Today).Return(int64(2000000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(300000), "USD", "RUB", test.Today).Return(int64(1500000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(30000000), "RUB", "RUB", test.Today).Return(int64(30000000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(10000000), "RUB", "RUB", test.Today).Return(int64(10000000), nil)
			},
		})

		// ACT
		resp := controller.ListLimits(context.Background(), request.ListLimits{
			User: test.User,
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
	t.Parallel()

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", test.SimpleError)
			},
		})

		// ACT
		resp := controller.SetLimit(context.Background(), request.SetLimit{
			User:     test.User,
			Value:    1000000,
			Category: "",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(false), resp)
	})

	t.Run("cannot set", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(test.CtxInterface), test.User, int64(1000000), "USD", "taxi").Return(test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(context.Background(), request.SetLimit{
			User:     test.User,
			Value:    1000000,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(false), resp)
	})

	t.Run("successful set", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(test.CtxInterface), test.User, int64(2000000), "USD", "taxi").Return(nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(context.Background(), request.SetLimit{
			User:     test.User,
			Value:    2000000,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(true), resp)
	})

	t.Run("cannot unset", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Unset(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "taxi").Return(test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(context.Background(), request.SetLimit{
			User:     test.User,
			Value:    0,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(false), resp)
	})

	t.Run("successful unset", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Unset(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "taxi").Return(nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.SetLimit(context.Background(), request.SetLimit{
			User:     test.User,
			Value:    0,
			Category: "taxi",
		})

		// ASSERT
		assert.Equal(t, response.SetLimit(true), resp)
	})
}

func Test_controller_AddExpense(t *testing.T) {
	t.Parallel()

	t.Run("not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(false)
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
			Amount:   10000,
			Category: "coffee",
		})

		// ASSERT
		assert.Empty(t, resp)
	})

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", test.SimpleError)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(20000), "USD", "coffee").Return(test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(25000), "USD", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "coffee").Return(types.LimitItem{}, test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(30000), "USD", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "coffee").Return(types.LimitItem{
					Total: 0,
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(35000), "EUR", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("EUR", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(35000), "EUR", "USD", test.Today).Return(int64(0), test.SimpleError)
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(2000000), "RUB", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
				m.EXPECT().Decrease(gomock.AssignableToTypeOf(test.CtxInterface), test.User, int64(40000), "coffee").Return(false, test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(2000000), "RUB", "USD", test.Today).Return(int64(40000), nil)
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(2000000), "RUB", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
				m.EXPECT().Decrease(gomock.AssignableToTypeOf(test.CtxInterface), test.User, int64(40000), "coffee").Return(true, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(2000000), "RUB", "USD", test.Today).Return(int64(40000), nil)
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
		controller := setupController(t, controllerMocksInitializer{
			expenser: func(m *mocks.MockExpenser) {
				m.EXPECT().AddExpense(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, int64(1000000), "RUB", "coffee").Return(nil)
			},
			limiter: func(m *mocks.Mocklimiter) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "coffee").Return(types.LimitItem{
					Total:    60000,
					Remains:  30000,
					Currency: "USD",
				}, nil)
				m.EXPECT().Decrease(gomock.AssignableToTypeOf(test.CtxInterface), test.User, int64(20000), "coffee").Return(false, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
			rater: func(m *mocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(1000000), "RUB", "USD", test.Today).Return(int64(20000), nil)
			},
		})

		// ACT
		resp := controller.AddExpense(context.Background(), request.AddExpense{
			User:     test.User,
			Date:     test.Today,
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
	t.Parallel()

	t.Run("no currency", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", test.SimpleError)
			},
		})

		// ACT
		resp := controller.GetReport(context.Background(), request.GetReport{
			User: test.User,
			From: test.Today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:  test.Today,
			Ready: false,
		}, resp)
	})

	t.Run("not ready", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			reporter: func(m *mocks.MockReporter) {
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, "RUB").Return(nil, ErrNotReady)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
		})

		// ACT
		resp := controller.GetReport(context.Background(), request.GetReport{
			User: test.User,
			From: test.Today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:     test.Today,
			Currency: "RUB",
			Ready:    false,
		}, resp)
	})

	t.Run("ready error", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		controller := setupController(t, controllerMocksInitializer{
			reporter: func(m *mocks.MockReporter) {
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, "RUB").Return(nil, test.SimpleError)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("RUB", nil)
			},
		})

		// ACT
		resp := controller.GetReport(context.Background(), request.GetReport{
			User: test.User,
			From: test.Today,
		})

		// ASSERT
		assert.Equal(t, response.GetReport{
			From:     test.Today,
			Currency: "RUB",
			Ready:    true,
		}, resp)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		// ARRANGE
		expectedResp := response.GetReport{
			From:     test.Today,
			Currency: "USD",
			Ready:    true,
			Data: map[string]int64{
				"coffee": 20000,
				"taxi":   130000,
			},
			Success: true,
		}

		controller := setupController(t, controllerMocksInitializer{
			reporter: func(m *mocks.MockReporter) {
				m.EXPECT().GetReport(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, "USD").Return(map[string]int64{
					"coffee": 20000,
					"taxi":   130000,
				}, nil)
			},
			currencyManager: func(m *mocks.MockcurrencyManager) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("USD", nil)
			},
		})

		// ACT
		resp := controller.GetReport(context.Background(), request.GetReport{
			User: test.User,
			From: test.Today,
		})

		// ASSERT
		assert.Equal(t, expectedResp, resp)
	})
}
