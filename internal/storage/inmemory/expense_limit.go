package inmemory

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryExpenseLimitStorage struct {
	data map[*types.User]map[string]types.LimitItem
}

func (s *inMemoryExpenseLimitStorage) Get(ctx context.Context, user *types.User, category string) (types.LimitItem, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryExpenseLimitStorage.Get")
	defer span.Finish()

	if limits, found := s.data[user]; found {
		if limit, found := limits[category]; found {
			return limit, found, nil
		}

		if limit, found := limits[""]; found {
			return limit, found, nil
		}
	}

	return types.LimitItem{}, false, nil
}

func (s *inMemoryExpenseLimitStorage) Set(ctx context.Context, user *types.User, total int64, currency, category string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryExpenseLimitStorage.Set")
	defer span.Finish()

	item := types.LimitItem{
		Total:    total,
		Remains:  total,
		Currency: currency,
	}

	if _, ok := s.data[user]; ok {
		s.data[user][category] = item
	} else {
		s.data[user] = map[string]types.LimitItem{category: item}
	}

	return nil
}

func (s *inMemoryExpenseLimitStorage) Decrease(ctx context.Context, user *types.User, value int64, category string) (bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryExpenseLimitStorage.Decrease")
	defer span.Finish()

	if _, ok := s.data[user]; !ok {
		return true, nil
	}

	item, ok := s.data[user][category]
	if !ok {
		category = ""
		item, ok = s.data[user][category]
	}
	if !ok {
		return true, nil
	}

	item.Remains -= value
	if item.Remains < 0 {
		item.Remains = 0
	}

	s.data[user][category] = item

	return item.Remains == 0, nil
}

func (s *inMemoryExpenseLimitStorage) Unset(ctx context.Context, user *types.User, category string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryExpenseLimitStorage.Unset")
	defer span.Finish()

	if _, ok := s.data[user]; ok {
		delete(s.data[user], category)
	}

	return nil
}

func (s *inMemoryExpenseLimitStorage) List(ctx context.Context, user *types.User) (map[string]types.LimitItem, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryExpenseLimitStorage.List")
	defer span.Finish()

	limits, found := s.data[user]

	return limits, found, nil
}
