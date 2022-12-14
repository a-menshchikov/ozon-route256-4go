package postgresql

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"go.uber.org/zap"
)

const (
	_defaultWaitTimeout = 5 * time.Second
)

type factory struct {
	pool *pgxpool.Pool
}

func NewFactory(ctx context.Context, dsn string, waitTimeout time.Duration, logger *zap.Logger) (*factory, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := wait(pool, ctx, waitTimeout, logger); err != nil {
		pool.Close()
		return nil, err
	}

	go func() {
		<-ctx.Done()
		pool.Close()
	}()

	prometheus.MustRegister(newExporter(pool, pool.Config().ConnConfig.Database))

	return &factory{
		pool: pool,
	}, nil
}

func wait(conn driver.Pinger, ctx context.Context, waitTimeout time.Duration, logger *zap.Logger) error {
	if waitTimeout == 0 {
		waitTimeout = _defaultWaitTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	if err := ping(ctx, conn, logger); err == nil {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := ping(ctx, conn, nil); err == nil {
				return nil
			}
		}
	}
}

func ping(ctx context.Context, conn driver.Pinger, logger *zap.Logger) error {
	err := conn.Ping(ctx)
	if err != nil {
		logger.Info("db ping", zap.Error(err))
	}

	return err
}

func (f *factory) CreateTelegramUserStorage() storage.TelegramUserStorage {
	return &pgTelegramUserStorage{
		pool: f.pool,
	}
}

func (f *factory) CreateExpenseStorage() storage.ExpenseStorage {
	return &pgExpenseStorage{
		pool: f.pool,
	}
}

func (f *factory) CreateCurrencyStorage() storage.CurrencyStorage {
	return &pgCurrencyStorage{
		pool: f.pool,
	}
}

func (f *factory) CreateCurrencyRatesStorage() storage.CurrencyRatesStorage {
	return &pgCurrencyRatesStorage{
		pool: f.pool,
	}
}

func (f *factory) CreateExpenseLimitStorage() storage.ExpenseLimitStorage {
	return &pgExpenseLimitStorage{
		pool: f.pool,
	}
}
