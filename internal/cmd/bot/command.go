package bot

import (
	"context"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/cache/redis"
	tgclient "gitlab.ozon.dev/almenschhikov/go-course-4/internal/clients/telegram"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/ctxkey"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/metrics"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/currency"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/currency/cbr"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/expense"
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
	_defaultMetricsPort = 8084
)

type (
	storageFactory interface {
		CreateTelegramUserStorage() storage.TelegramUserStorage
		CreateExpenseStorage() storage.ExpenseStorage
		CreateExpenseLimitStorage() storage.ExpenseLimitStorage
		CreateCurrencyStorage() storage.CurrencyStorage
		CreateCurrencyRatesStorage() storage.CurrencyRatesStorage
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

		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			logger, err := utils.NewLogger(logLevel, logDevel)
			if err != nil {
				return errors.Wrap(err, "logger init failed")
			}

			cmd.SetContext(context.WithValue(cmd.Context(), ctxkey.Logger, logger))
			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			logger := ctx.Value(ctxkey.Logger).(*zap.Logger)

			closer, err := utils.InitTracing(serviceName)
			if err != nil {
				return errors.Wrap(err, "tracer init failed")
			}
			defer func(c io.Closer, l *zap.Logger) {
				if err := c.Close(); err != nil {
					l.Error("cannot close tracer", zap.Error(err))
				}
			}(closer, logger)

			cfg, err := config.NewConfig(configPath)
			if err != nil {
				return errors.Wrap(err, "config init failed")
			}

			g, ctx := errgroup.WithContext(ctx)

			factory, err := newStorageFactory(ctx, cfg.Storage, logger)
			if err != nil {
				return errors.Wrap(err, "storage factory init failed")
			}

			ratesStorage := factory.CreateCurrencyRatesStorage()
			if cfg.Cache.Rates.Driver != "" {
				if ratesStorage, err = newCurrencyRatesCache(ratesStorage, cfg.Cache.Rates, logger); err != nil {
					logger.Error("rates cache init failed", zap.Error(err))
				}
			}

			metricsServer := metrics.NewServer(uint16(metricsPort), logger)
			g.Go(func() error {
				return metricsServer.Run(ctx)
			})

			cbrGateway := cbr.NewCbrGateway(http.DefaultClient)
			rater := currency.NewRater(cfg.Currency, ratesStorage, cbrGateway, logger)
			g.Go(func() error {
				return rater.Run(ctx)
			})

			reportsListener, err := reports.NewListener(cfg.Reports.Grpc, logger)
			if err != nil {
				return errors.Wrap(err, "reports listener init failed")
			}
			defer reportsListener.Close()

			reportsListener.Run()

			reportsProducer, err := reports.NewProducer(cfg.Reports.Kafka, logger)
			if err != nil {
				return errors.Wrap(err, "reports message producer init failed")
			}
			defer reportsProducer.Close()

			var (
				expenser model.Expenser = expense.NewExpenser(factory.CreateExpenseStorage())
				reporter model.Reporter = expense.NewReporter(cfg.Reports.Kafka.Timeout, reportsProducer, reportsListener, logger)
			)
			if cfg.Cache.Reporter.Driver != "" {
				if expenser, reporter, err = newReportCache(expenser, reporter, cfg.Cache.Reporter, logger); err != nil {
					logger.Error("reports cache init failed", zap.Error(err))
				}
			}

			tgClient, err := newTelegramClient(cfg.Client.Telegram.Token, factory.CreateTelegramUserStorage(), logger)
			if err != nil {
				return errors.Wrap(err, "telegram client init failed")
			}

			limiter := expense.NewLimiter(factory.CreateExpenseLimitStorage())
			currencyManager := currency.NewCurrencyManager(cfg.Currency, factory.CreateCurrencyStorage())
			finAssist := model.NewController(expenser, reporter, limiter, currencyManager, rater, logger)

			tgClient.RegisterController(finAssist)
			g.Go(func() error {
				return tgClient.ListenUpdates(ctx)
			})

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
	c.PersistentFlags().StringVar(&serviceName, "service", "finassist_bot", "service name for tracing")

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
