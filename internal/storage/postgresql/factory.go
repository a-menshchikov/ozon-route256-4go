package postgresql

import (
	"context"
	"database/sql/driver"
	"log"
	"time"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
)

const (
	_defaultWaitTimeout = 5 * time.Second
)

type factory struct {
	ctx context.Context
	db  pgxtype.Querier
}

func NewFactory(ctx context.Context, dsn string, waitTimeout time.Duration) (*factory, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := wait(pool, ctx, waitTimeout); err != nil {
		pool.Close()
		return nil, err
	}

	go func() {
		<-ctx.Done()
		pool.Close()
	}()

	return &factory{
		ctx: ctx,
		db:  pool,
	}, nil
}

func wait(conn driver.Pinger, ctx context.Context, waitTimeout time.Duration) error {
	if waitTimeout == 0 {
		waitTimeout = _defaultWaitTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	if err := ping(ctx, conn); err == nil {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := ping(ctx, conn); err == nil {
				return nil
			}
		}
	}
}

func ping(ctx context.Context, conn driver.Pinger) error {
	err := conn.Ping(ctx)
	if err != nil {
		log.Println("db wait:", err)
	}

	return err
}

func (f *factory) CreateTelegramUserStorage() storage.TelegramUserStorage {
	return &pgTelegramUserStorage{
		ctx: f.ctx,
		db:  f.db,
	}
}

func (f *factory) CreateExpenseStorage() storage.ExpenseStorage {
	return &pgExpenseStorage{
		ctx: f.ctx,
		db:  f.db,
	}
}

func (f *factory) CreateCurrencyStorage() storage.CurrencyStorage {
	return &pgCurrencyStorage{
		ctx: f.ctx,
		db:  f.db,
	}
}

func (f *factory) CreateRatesStorage() storage.CurrencyRatesStorage {
	return &pgCurrencyRatesStorage{
		ctx: f.ctx,
		db:  f.db,
	}
}

func (f *factory) CreateLimitStorage() storage.ExpenseLimitStorage {
	return &pgExpenseLimitStorage{
		ctx: f.ctx,
		db:  f.db,
	}
}
