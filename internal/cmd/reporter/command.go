package reporter

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/cache/redis"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/metrics"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/currency"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/expense/reports"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage/inmemory"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage/postgresql"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

const (
	_defaultMetricsPort = 8085
)

type (
	storageFactory interface {
		CreateExpenseStorage() storage.ExpenseStorage
		CreateRatesStorage() storage.CurrencyRatesStorage
	}
)

func NewCommand(name, version string) *cobra.Command {
	var (
		configPath  string
		logLevel    string
		logDevel    bool
		metricsPort int
		serviceName string
	)

	c := &cobra.Command{
		Use:           name,
		Short:         "Financial Assistant bot",
		Example:       name + " --config=.data/config.yaml",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,

		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := utils.NewLogger(logLevel, logDevel)
			if err != nil {
				return errors.Wrap(err, "logger init failed")
			}

			if err := utils.InitTracing(serviceName); err != nil {
				return errors.Wrap(err, "tracer init failed")
			}

			cfg, err := config.NewConfig(configPath)
			if err != nil {
				return errors.Wrap(err, "config init failed")
			}

			g, ctx := errgroup.WithContext(cmd.Context())

			storageFactory, err := newStorageFactory(ctx, cfg.Storage, logger)
			if err != nil {
				return errors.Wrap(err, "storage factory init failed")
			}

			ratesStorage := storageFactory.CreateRatesStorage()
			if cfg.Cache.Rates.Driver != "" {
				if ratesStorage, err = newCurrencyRatesCache(ratesStorage, cfg.Cache.Rates, logger); err != nil {
					logger.Error("rates cache init failed", zap.Error(err))
				}
			}

			metricsServer := metrics.NewServer(uint16(metricsPort), logger)

			g.Go(func() error { return metricsServer.Run(ctx) })

			expenseStorage := storageFactory.CreateExpenseStorage()
			rater := currency.NewRater(cfg.Currency, ratesStorage, nil, logger)
			consumer, err := reports.NewConsumer(cfg.Reports.Kafka, cfg.Reports.Grpc, expenseStorage, rater, logger)
			if err != nil {
				return errors.Wrap(err, "reports consumer init failed")
			}

			g.Go(func() error { return consumer.Run(ctx) })

			if err := g.Wait(); err != nil {
				return err
			}

			return nil
		},
	}

	c.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config path")
	c.PersistentFlags().StringVar(&logLevel, "log-level", zapcore.InfoLevel.String(), "debug | info | warn | error | dpanic | panic | fatal")
	c.PersistentFlags().BoolVar(&logDevel, "log-devel", false, "use development logging")
	c.PersistentFlags().IntVar(&metricsPort, "metrics-port", _defaultMetricsPort, "http port for metrics collecting")
	c.PersistentFlags().StringVar(&serviceName, "service", "finassist_reporter", "service name for tracing")

	_ = c.MarkFlagRequired("config")

	return c
}

func newStorageFactory(ctx context.Context, cfg config.StorageConfig, logger *zap.Logger) (storageFactory, error) {
	switch cfg.Driver {
	case config.InMemoryDriver:
		return inmemory.NewFactory(), nil

	case config.PostgreSQLDriver:
		return postgresql.NewFactory(ctx, cfg.Dsn, cfg.WaitTimeout, logger)
	}

	return nil, errors.New("unknown storage driver")
}

func newCurrencyRatesCache(storage storage.CurrencyRatesStorage, cfg config.CacheSectionConfig, logger *zap.Logger) (storage.CurrencyRatesStorage, error) {
	switch cfg.Driver {
	case config.RedisDriver:
		return redis.NewCurrencyRatesCache(storage, cfg.Dsn, logger)
	}

	return storage, errors.New("unknown rates cache driver")
}
