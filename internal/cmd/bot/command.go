package bot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	jaegierconfig "github.com/uber/jaeger-client-go/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/cache/redis"
	tgclient "gitlab.ozon.dev/almenschhikov/go-course-4/internal/clients/telegram"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/metrics"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/currency"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/currency/cbr"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/expense"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage/inmemory"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage/postgresql"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

const (
	_defaultMetricsPort = 8084
)

type (
	storageFactory interface {
		CreateTelegramUserStorage() storage.TelegramUserStorage
		CreateExpenseStorage() storage.ExpenseStorage
		CreateLimitStorage() storage.ExpenseLimitStorage
		CreateCurrencyStorage() storage.CurrencyStorage
		CreateRatesStorage() storage.CurrencyRatesStorage
	}

	client interface {
		RegisterController(model.Controller)
		ListenUpdates(ctx context.Context) error
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
			logger, err := newLogger(logLevel, logDevel)
			if err != nil {
				return errors.Wrap(err, "logger init failed")
			}

			if err := initTracing(serviceName); err != nil {
				return errors.Wrap(err, "tracer init failed")
			}

			cfg, err := config.New(configPath)
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
			rater := currency.NewRater(
				cfg.Currency,
				ratesStorage,
				cbr.NewCbrGateway(http.DefaultClient),
				logger,
			)

			g.Go(func() error { return metricsServer.Run(ctx) })
			g.Go(func() error { return rater.Run(ctx) })

			expenseStorage := storageFactory.CreateExpenseStorage()
			var (
				expenser model.Expenser = expense.NewExpenser(expenseStorage)
				reporter model.Reporter = expense.NewReporter(expenseStorage, rater, logger)
			)
			if cfg.Cache.Reporter.Driver != "" {
				if expenser, reporter, err = newReportCache(expenser, reporter, cfg.Cache.Reporter, logger); err != nil {
					logger.Error("reports cache init failed", zap.Error(err))
				}
			}

			finAssist := model.NewController(
				expenser,
				reporter,
				expense.NewLimiter(storageFactory.CreateLimitStorage()),
				currency.NewCurrencyManager(cfg.Currency, storageFactory.CreateCurrencyStorage()),
				rater,
				logger,
			)

			tgClient, err := newTelegramClient(cfg.Client.Telegram.Token, storageFactory.CreateTelegramUserStorage(), logger)
			if err != nil {
				return errors.Wrap(err, "telegram client init failed")
			}

			tgClient.RegisterController(finAssist)
			g.Go(func() error { return tgClient.ListenUpdates(ctx) })

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
	c.PersistentFlags().StringVar(&serviceName, "service", "finassist", "service name for tracing")

	_ = c.MarkFlagRequired("config")

	return c
}

func newLogger(logLevel string, developerMode bool) (*zap.Logger, error) {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		return nil, err
	}

	var cfg zap.Config
	if developerMode {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return logger, nil
}

func initTracing(serviceName string) error {
	cfg := jaegierconfig.Configuration{
		Sampler: &jaegierconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	_, err := cfg.InitGlobalTracer(serviceName)
	return err
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

func newReportCache(expenser model.Expenser, reporter model.Reporter, cfg config.CacheSectionConfig, logger *zap.Logger) (model.Expenser, model.Reporter, error) {
	switch cfg.Driver {
	case config.RedisDriver:
		cache, err := redis.NewReportCache(expenser, reporter, cfg.Dsn, logger)
		return cache, cache, err
	}

	return expenser, reporter, errors.New("unknown report cache driver")
}

func newTelegramClient(token string, s storage.TelegramUserStorage, l *zap.Logger) (client, error) {
	return tgclient.NewClient(token, s, l)
}
