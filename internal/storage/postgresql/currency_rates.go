package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type pgCurrencyRatesStorage struct {
	ctx context.Context
	db  pgxtype.Querier
}

func (s *pgCurrencyRatesStorage) Get(currency string, date time.Time) (int64, bool, error) {
	var rate int64

	err := s.db.QueryRow(
		s.ctx,
		`with date_filter as (select $1 code, $2::date date),
              rates as (select * from rates where code = $1),
              result as (select date,
                                coalesce(
                                  rate,
                                  lead(rate, 1) over (order by date),
                                  lag(rate, 1) over (order by date)
                                ) rt
                         from rates
                                full join date_filter using (code, date)
                         where code = $1)
         select rt
         from result`,
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

func (s *pgCurrencyRatesStorage) Add(currency string, date time.Time, rate int64) error {
	_, err := s.db.Exec(
		s.ctx,
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
