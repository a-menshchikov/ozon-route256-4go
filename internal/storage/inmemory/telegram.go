package inmemory

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryTelegramUserStorage struct {
	data map[int64]*types.User
}

func (s *inMemoryTelegramUserStorage) Add(ctx context.Context, tgUserID int64) (*types.User, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryTelegramUserStorage.Add")
	defer span.Finish()

	if _, ok := s.data[tgUserID]; ok {
		return nil, errors.New("user already exists")
	}

	user := types.User(tgUserID)
	s.data[tgUserID] = &user

	return s.data[tgUserID], nil
}

func (s *inMemoryTelegramUserStorage) FetchByID(ctx context.Context, tgUserID int64) (*types.User, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "inMemoryTelegramUserStorage.FetchByID")
	defer span.Finish()

	if user, ok := s.data[tgUserID]; ok {
		return user, nil
	}

	return nil, errors.New("user not found")
}
