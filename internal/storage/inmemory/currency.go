package inmemory

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryCurrencyStorage struct {
	data map[*types.User]string
}

func (s *inMemoryCurrencyStorage) Get(ctx context.Context, user *types.User) (string, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryCurrencyStorage.Get")
	defer span.Finish()

	currency, found := s.data[user]

	return currency, found, nil
}

func (s *inMemoryCurrencyStorage) Set(ctx context.Context, user *types.User, value string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryCurrencyStorage.Set")
	defer span.Finish()

	s.data[user] = value

	return nil
}
