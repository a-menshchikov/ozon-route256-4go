package model

import (
	"context"
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/request"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/response"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type (
	Controller interface {
		ListCurrencies(ctx context.Context, req request.ListCurrencies) response.ListCurrencies
		SetCurrency(ctx context.Context, req request.SetCurrency) response.SetCurrency

		ListLimits(ctx context.Context, req request.ListLimits) response.ListLimits
		SetLimit(ctx context.Context, req request.SetLimit) response.SetLimit

		AddExpense(ctx context.Context, req request.AddExpense) response.AddExpense

		GetReport(ctx context.Context, req request.GetReport) response.GetReport
	}

	Expenser interface {
		AddExpense(ctx context.Context, user *types.User, date time.Time, amount int64, currency, category string) error
	}

	Reporter interface {
		GetReport(ctx context.Context, user *types.User, from time.Time, currency string) (map[string]int64, error)
	}

	Rater interface {
		Run(ctx context.Context) error
		TryAcquireExchange() bool
		ReleaseExchange()
		Exchange(ctx context.Context, value int64, from, to string, date time.Time) (int64, error)
	}

	limiter interface {
		Get(ctx context.Context, user *types.User, category string) (types.LimitItem, error)
		Set(ctx context.Context, user *types.User, limit int64, currency, category string) error
		Decrease(ctx context.Context, user *types.User, value int64, category string) (bool, error)
		Unset(ctx context.Context, user *types.User, category string) error
		List(ctx context.Context, user *types.User) (map[string]types.LimitItem, error)
	}

	currencyManager interface {
		Get(ctx context.Context, user *types.User) (string, error)
		Set(ctx context.Context, user *types.User, currency string) error
		ListCurrenciesCodesWithFlags() []string
	}
)
