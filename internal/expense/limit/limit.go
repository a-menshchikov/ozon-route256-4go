package limit

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
)

type limiter struct {
	storage storage.ExpenseLimitStorage
}

func NewLimiter(s storage.ExpenseLimitStorage) *limiter {
	return &limiter{
		storage: s,
	}
}

func (l *limiter) Get(user *types.User, category string) (types.LimitItem, error) {
	limit, found, err := l.storage.Get(user, category)
	if err != nil {
		return types.LimitItem{}, errors.Wrap(err, "ExpenseLimitStorage.Get")
	}

	if found {
		return limit, nil
	}

	return types.LimitItem{}, nil
}

func (l *limiter) Set(user *types.User, limit int64, currency, category string) error {
	if limit < 0 {
		return errors.New("negative limit")
	}

	return l.storage.Set(user, limit, currency, category)
}

func (l *limiter) Decrease(user *types.User, value int64, category string) (bool, error) {
	return l.storage.Decrease(user, value, category)
}

func (l *limiter) Unset(user *types.User, category string) error {
	return l.storage.Unset(user, category)
}

func (l *limiter) List(user *types.User) (map[string]types.LimitItem, error) {
	list, found, err := l.storage.List(user)
	if err != nil {
		return nil, errors.Wrap(err, "ExpenseLimitStorage.Report")
	}

	if found {
		return list, nil
	}

	return nil, nil
}
