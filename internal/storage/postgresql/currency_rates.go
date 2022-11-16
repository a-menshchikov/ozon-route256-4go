package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

type pgCurrencyRatesStorage struct {
	pool *pgxpool.Pool
}

func (s *pgCurrencyRatesStorage) Get(ctx context.Context, currency string, date time.Time) (int64, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgCurrencyRatesStorage.Get")
	defer span.Finish()

	var rate int64

	err := s.pool.QueryRow(
		ctx,
		`with date_filter as (select $1 code, $2::date date),
              rates as (select * from rates where code = $1),
              result as (select date,
                                coalesce(
                                  rate,
                                  lead(rate, 1) over (order by date),
                                  lag(rate, 1) over (order by date)
                                ) as rate
                         from date_filter
                                full join rates using (code, date)
                         where code = $1)
         select rate
         from result
         where date = $2::date
           and rate is not null`,
		currency, // $1
		date,     // $2
	).Scan(&rate)
	if err == pgx.ErrNoRows {
		return 0, false, nil
	} else if err != nil {
		return 0, false, errors.Wrap(err, "select rate")
	}

	return rate, true, nil
}

func (s *pgCurrencyRatesStorage) Add(ctx context.Context, currency string, date time.Time, rate int64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgCurrencyRatesStorage.Add")
	defer span.Finish()

	_, err := s.pool.Exec(
		ctx,
		`insert into rates (code, date, rate)
         values ($1, $2, $3)
           on conflict (code, date)
             do update set rate = EXCLUDED.rate`,
		currency, // $1
		date,     // $2
		rate,     // $3
	)
	if err != nil {
		return errors.Wrap(err, "add rate")
	}

	return nil
}
