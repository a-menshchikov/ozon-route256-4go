package expense

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

type expenser struct {
	storage storage.ExpenseStorage
}

func NewExpenser(s storage.ExpenseStorage) *expenser {
	return &expenser{
		storage: s,
	}
}

func (e *expenser) Add(user *types.User, date time.Time, amount int64, currency, category string) error {
	if amount < 0 {
		return errors.New("сумма трат должна быть положительным числом")
	}

	if date.After(utils.TruncateToDate(time.Now())) {
		return errors.New("траты из будущего не поддерживаются")
	}

	return e.storage.Add(
		user,
		types.ExpenseItem{
			Date:     date,
			Amount:   amount,
			Currency: currency,
		},
		category,
	)
}

func (e *expenser) Report(user *types.User, from time.Time) (map[string][]types.ExpenseItem, error) {
	data, err := e.storage.List(user, from)
	if err != nil {
		return nil, errors.Wrap(err, "ExpenseStorage.Report")
	}

	return data, nil
}
