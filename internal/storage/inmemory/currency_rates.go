package inmemory

import (
	"time"
)

type inMemoryCurrencyRatesStorage struct {
	data map[string]int64
}

func (s *inMemoryCurrencyRatesStorage) Get(currency string, _ time.Time) (int64, bool, error) {
	rate, found := s.data[currency]

	return rate, found, nil
}

func (s *inMemoryCurrencyRatesStorage) Add(currency string, rate int64, _ time.Time) error {
	s.data[currency] = rate

	return nil
}
