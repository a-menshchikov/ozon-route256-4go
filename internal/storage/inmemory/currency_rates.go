package inmemory

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
)

type inMemoryCurrencyRatesStorage struct {
	data map[string]int64
}

func (s *inMemoryCurrencyRatesStorage) Get(ctx context.Context, currency string, _ time.Time) (int64, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryCurrencyRatesStorage.Get")
	defer span.Finish()

	rate, found := s.data[currency]

	return rate, found, nil
}

func (s *inMemoryCurrencyRatesStorage) Add(ctx context.Context, currency string, _ time.Time, rate int64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryCurrencyRatesStorage.Add")
	defer span.Finish()

	s.data[currency] = rate

	return nil
}
