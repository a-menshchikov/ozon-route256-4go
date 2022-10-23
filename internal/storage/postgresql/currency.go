package postgresql

import (
	"context"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type pgCurrencyStorage struct {
	ctx context.Context
	db  pgxtype.Querier
}

func (s *pgCurrencyStorage) Get(user *types.User) (string, bool, error) {
	var value string

	err := s.db.QueryRow(
		s.ctx,
		`select code
         from currencies
         where user_id = $1`,
		user, // $1
	).Scan(&value)
	if err == pgx.ErrNoRows {
		return "", false, nil
	} else if err != nil {
		return "", false, errors.Wrap(err, "select currency")
	}

	return value, true, nil
}

func (s *pgCurrencyStorage) Set(user *types.User, value string) error {
	_, err := s.db.Exec(
		s.ctx,
		`insert into currencies (user_id, code)
         values ($1, $2)
           on conflict (user_id)
             do update set code = excluded.code`,
		user,  // $1
		value, // $2
	)
	if err != nil {
		return errors.Wrap(err, "upsert currency")
	}

	return nil
}
