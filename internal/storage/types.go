package storage

import (
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type TelegramUserStorage interface {
	Add(tgUserID int64) (*types.User, error)
	FetchByID(tgUserID int64) (*types.User, error)
}

type ExpenseStorage interface {
	Add(user *types.User, item types.ExpenseItem, category string) error
	List(user *types.User, from time.Time) (map[string][]types.ExpenseItem, error)
}

type ExpenseLimitStorage interface {
	Get(user *types.User, category string) (types.LimitItem, bool, error)
	Set(user *types.User, value int64, currency, category string) error
	Decrease(user *types.User, value int64, category string) (bool, error)
	Unset(user *types.User, category string) error
	List(user *types.User) (map[string]types.LimitItem, bool, error)
}

type CurrencyStorage interface {
	Get(user *types.User) (string, bool, error)
	Set(user *types.User, value string) error
}

type CurrencyRatesStorage interface {
	Get(currency string, date time.Time) (int64, bool, error)
	Add(currency string, rate int64, date time.Time) error
}
