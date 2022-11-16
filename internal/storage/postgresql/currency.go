package postgresql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type pgCurrencyStorage struct {
	pool *pgxpool.Pool
}

func (s *pgCurrencyStorage) Get(ctx context.Context, user *types.User) (string, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgCurrencyStorage.Get")
	defer span.Finish()

	var value string

	err := s.pool.QueryRow(
		ctx,
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

func (s *pgCurrencyStorage) Set(ctx context.Context, user *types.User, value string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgCurrencyStorage.Set")
	defer span.Finish()

	_, err := s.pool.Exec(
		ctx,
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
