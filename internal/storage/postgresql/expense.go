package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type pgExpenseStorage struct {
	pool *pgxpool.Pool
}

func (s *pgExpenseStorage) Add(ctx context.Context, user *types.User, item types.ExpenseItem, category string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseStorage.Add")
	defer span.Finish()

	_, err := s.pool.Exec(
		ctx,
		`insert into expenses (user_id, date, amount, currency_code, category)
         values ($1, $2, $3, $4, $5)`,
		user,          // $1
		item.Date,     // $2
		item.Amount,   // $3
		item.Currency, // $4
		category,      // $5
	)
	if err != nil {
		return errors.Wrap(err, "insert expense")
	}

	return nil
}

func (s *pgExpenseStorage) List(ctx context.Context, user *types.User, from time.Time) (map[string][]types.ExpenseItem, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pgExpenseStorage.List")
	defer span.Finish()

	rows, err := s.pool.Query(
		ctx,
		`select category, date, amount, currency_code
         from expenses
         where user_id = $1
           and date >= $2`,
		user, // $1
		from, // $2
	)
	if err != nil {
		return nil, errors.Wrap(err, "select expenses")
	}

	list := make(map[string][]types.ExpenseItem)

	var (
		category string
		date     time.Time
		amount   int64
		currency string
	)
	for rows.Next() {
		if err := rows.Scan(&category, &date, &amount, &currency); err != nil {
			return nil, errors.Wrap(err, "scan selected expenses")
		}

		list[category] = append(list[category], types.ExpenseItem{
			Date:     date,
			Amount:   amount,
			Currency: currency,
		})
	}

	return list, nil
}
