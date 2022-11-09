package postgresql

import (
	"context"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type pgExpenseLimitStorage struct {
	db pgxtype.Querier
}

func (s *pgExpenseLimitStorage) Get(ctx context.Context, user *types.User, category string) (types.LimitItem, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseLimitStorage.Get")
	defer span.Finish()

	var (
		total    int64
		remains  int64
		currency string
	)

	err := s.db.QueryRow(
		ctx,
		`select total,
                remains,
                currency_code
         from limits
         where user_id = $1
           and category in ($2, '')
         order by category desc
         limit 1`,
		user,     // $1
		category, // $2
	).Scan(&total, &remains, &currency)
	if err == pgx.ErrNoRows {
		return types.LimitItem{}, false, nil
	} else if err != nil {
		return types.LimitItem{}, false, errors.Wrap(err, "select limit")
	}

	return types.LimitItem{
		Total:    total,
		Remains:  remains,
		Currency: currency,
	}, true, nil
}

func (s *pgExpenseLimitStorage) Set(ctx context.Context, user *types.User, total int64, currency, category string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseLimitStorage.Set")
	defer span.Finish()

	_, err := s.db.Exec(
		ctx,
		`insert into limits (user_id, category, total, remains, currency_code)
		 values ($1, $4, $2, $2, $3)
		   on conflict (user_id, category)
		     do update set total = excluded.total,
		                   remains = excluded.remains,
		                   currency_code = excluded.currency_code`,
		user,     // $1
		total,    // $2
		currency, // $3
		category, // $4
	)
	if err != nil {
		return errors.Wrap(err, "upsert limit")
	}

	return nil
}

func (s *pgExpenseLimitStorage) Decrease(ctx context.Context, user *types.User, value int64, category string) (bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseLimitStorage.Decrease")
	defer span.Finish()

	var limitReached bool

	err := s.db.QueryRow(
		ctx,
		`with "limit" as (
           select user_id, category
           from limits
           where user_id = $1
             and category in ($3, '')
           order by category desc
           limit 1
         )
         update limits set remains = case when remains - $2 > 0 then remains - $2 else 0 end
           from "limit"
           where limits.user_id = "limit".user_id
             and limits.category = "limit".category
         returning remains = 0`,
		user,     // $1
		value,    // $2
		category, // $3
	).Scan(&limitReached)
	if err != nil {
		return false, errors.Wrap(err, "decrease limit")
	}

	return limitReached, nil
}

func (s *pgExpenseLimitStorage) Unset(ctx context.Context, user *types.User, category string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseLimitStorage.Unset")
	defer span.Finish()

	_, err := s.db.Exec(
		ctx,
		`delete from limits
         where user_id = $1
           and category = $2`,
		user,     // $1
		category, // $2
	)
	if err != nil {
		return errors.Wrap(err, "delete limit")
	}

	return nil
}

func (s *pgExpenseLimitStorage) List(ctx context.Context, user *types.User) (map[string]types.LimitItem, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseLimitStorage.List")
	defer span.Finish()

	rows, err := s.db.Query(
		ctx,
		`select category,
                total,
                remains,
                currency_code
         from limits
         where user_id = $1`,
		user, // $1
	)
	if err != nil {
		return nil, false, errors.Wrap(err, "select limits")
	}

	list := make(map[string]types.LimitItem)
	found := false

	var (
		category     string
		total        int64
		remains      int64
		currencyCode string
	)
	for rows.Next() {
		found = true

		if err := rows.Scan(&category, &total, &remains, &currencyCode); err != nil {
			return nil, true, errors.Wrap(err, "scan selected limits")
		}

		list[category] = types.LimitItem{
			Total:    total,
			Remains:  remains,
			Currency: currencyCode,
		}
	}

	return list, found, nil
}
