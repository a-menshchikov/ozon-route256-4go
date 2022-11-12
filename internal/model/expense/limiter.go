package expense

import (
	"context"

	"github.com/opentracing/opentracing-go"
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

func (l *limiter) Get(ctx context.Context, user *types.User, category string) (types.LimitItem, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "limiter.Get", opentracing.Tags{
		"user":     *user,
		"category": category,
	})
	defer span.Finish()

	limit, found, err := l.storage.Get(ctx, user, category)
	if err != nil {
		return types.LimitItem{}, errors.Wrap(err, "ExpenseLimitStorage.Get")
	}

	if found {
		return limit, nil
	}

	return types.LimitItem{}, nil
}

func (l *limiter) Set(ctx context.Context, user *types.User, limit int64, currency, category string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "limiter.Set", opentracing.Tags{
		"user":     *user,
		"limit":    limit,
		"category": category,
	})
	defer span.Finish()

	if limit < 0 {
		return errors.New("negative limit")
	}

	return l.storage.Set(ctx, user, limit, currency, category)
}

func (l *limiter) Decrease(ctx context.Context, user *types.User, value int64, category string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "limiter.Decrease", opentracing.Tags{
		"user":     *user,
		"value":    value,
		"category": category,
	})
	defer span.Finish()

	return l.storage.Decrease(ctx, user, value, category)
}

func (l *limiter) Unset(ctx context.Context, user *types.User, category string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "limiter.Unset", opentracing.Tags{"user": *user})
	defer span.Finish()

	return l.storage.Unset(ctx, user, category)
}

func (l *limiter) List(ctx context.Context, user *types.User) (map[string]types.LimitItem, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "limiter.List", opentracing.Tags{"user": *user})
	defer span.Finish()

	list, found, err := l.storage.List(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "ExpenseLimitStorage.List")
	}

	if found {
		return list, nil
	}

	return nil, nil
}
