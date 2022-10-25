package postgresql

import (
	"context"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type pgTelegramUserStorage struct {
	ctx context.Context
	db  pgxtype.Querier
}

func (s *pgTelegramUserStorage) Add(tgUserID int64) (*types.User, error) {
	var userId int64
	err := s.db.QueryRow(
		s.ctx,
		`insert into users
         values (default)
           returning id`,
	).Scan(&userId)
	if err != nil {
		return nil, errors.Wrap(err, "insert new user")
	}

	_, err = s.db.Exec(
		s.ctx,
		`insert into tg_users (id, user_id)
         values ($1, $2)`,
		tgUserID, // $1
		userId,   // $2
	)
	if err != nil {
		return nil, errors.Wrap(err, "insert new telegram user")
	}

	user := types.User(userId)
	return &user, nil
}

func (s *pgTelegramUserStorage) FetchByID(tgUserID int64) (*types.User, error) {
	var userId int64

	err := s.db.QueryRow(
		s.ctx,
		`select user_id
         from tg_users
         where id = $1`,
		tgUserID, // $1
	).Scan(&userId)
	if err != nil {
		return nil, errors.Wrap(err, "select telegram user")
	}

	user := types.User(userId)
	return &user, nil
}
