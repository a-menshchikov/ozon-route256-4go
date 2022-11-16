package expense

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
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

func (e *expenser) AddExpense(ctx context.Context, user *types.User, date time.Time, amount int64, currency, category string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "expenser.Add", opentracing.Tags{
		"user":     *user,
		"date":     date,
		"amount":   amount,
		"currency": currency,
		"category": category,
	})
	defer span.Finish()

	if amount < 0 {
		return errors.New("сумма трат должна быть положительным числом")
	}

	if date.After(utils.TruncateToDate(time.Now())) {
		return errors.New("траты из будущего не поддерживаются")
	}

	return e.storage.Add(
		ctx,
		user,
		types.ExpenseItem{
			Date:     date,
			Amount:   amount,
			Currency: currency,
		},
		category,
	)
}
