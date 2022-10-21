package inmemory

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryCurrencyStorage struct {
	data map[*types.User]string
}

func (s *inMemoryCurrencyStorage) Get(user *types.User) (string, bool, error) {
	currency, found := s.data[user]

	return currency, found, nil
}

func (s *inMemoryCurrencyStorage) Set(user *types.User, value string) error {
	s.data[user] = value

	return nil
}
