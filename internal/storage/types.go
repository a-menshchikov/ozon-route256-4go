package storage

import (
	"context"
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type (
	TelegramUserStorage interface {
		Add(ctx context.Context, tgUserID int64) (*types.User, error)
		FetchByID(ctx context.Context, tgUserID int64) (*types.User, error)
	}

	ExpenseStorage interface {
		Add(ctx context.Context, user *types.User, item types.ExpenseItem, category string) error
		List(ctx context.Context, user *types.User, from time.Time) (map[string][]types.ExpenseItem, error)
	}

	ExpenseLimitStorage interface {
		Get(ctx context.Context, user *types.User, category string) (types.LimitItem, bool, error)
		Set(ctx context.Context, user *types.User, total int64, currency, category string) error
		Decrease(ctx context.Context, user *types.User, value int64, category string) (bool, error)
		Unset(ctx context.Context, user *types.User, category string) error
		List(ctx context.Context, user *types.User) (map[string]types.LimitItem, bool, error)
	}

	CurrencyStorage interface {
		Get(ctx context.Context, user *types.User) (string, bool, error)
		Set(ctx context.Context, user *types.User, value string) error
	}

	CurrencyRatesStorage interface {
		Get(ctx context.Context, currency string, date time.Time) (int64, bool, error)
		Add(ctx context.Context, currency string, date time.Time, rate int64) error
	}
)
