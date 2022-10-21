package inmemory

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryTelegramUserStorage struct {
	data map[int64]*types.User
}

func (s *inMemoryTelegramUserStorage) Add(tgUserID int64) (*types.User, error) {
	if _, ok := s.data[tgUserID]; ok {
		return nil, errors.New("user already exists")
	}

	s.data[tgUserID] = &([]types.User{types.User(tgUserID)}[0])

	return s.data[tgUserID], nil
}

func (s *inMemoryTelegramUserStorage) FetchByID(tgUserID int64) (*types.User, error) {
	if user, ok := s.data[tgUserID]; ok {
		return user, nil
	}

	return nil, errors.New("user not found")
}
