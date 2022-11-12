package expense

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type reporter struct {
	storage storage.ExpenseStorage
	rater   model.Rater
	logger  *zap.Logger
}

func NewReporter(s storage.ExpenseStorage, r model.Rater, l *zap.Logger) *reporter {
	return &reporter{
		storage: s,
		rater:   r,
		logger:  l,
	}
}

func (r *reporter) GetReport(ctx context.Context, user *types.User, from time.Time, currency string) (map[string]int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reporter.GetReport", opentracing.Tags{
		"user":     *user,
		"from":     from,
		"currency": currency,
	})
	defer span.Finish()

	if !r.rater.TryAcquireExchange() {
		return nil, model.ErrNotReady
	}
	defer r.rater.ReleaseExchange()

	expenses, err := r.storage.List(ctx, user, from)
	if err != nil {
		return nil, errors.Wrap(err, "ExpenseStorage.List")
	}

	data := make(map[string]int64)

	for category := range expenses {
		for _, item := range expenses[category] {
			amount, err := r.rater.Exchange(ctx, item.Amount, item.Currency, currency, item.Date)
			if err != nil {
				r.logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", item.Currency), zap.String("to", currency))

				return nil, errors.Wrap(err, "Rater.Exchange")
			}

			data[category] += amount
		}
	}

	return data, nil
}
