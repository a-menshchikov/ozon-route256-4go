package bot

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	jaegierconfig "github.com/uber/jaeger-client-go/config"
	tgclient "gitlab.ozon.dev/almenschhikov/go-course-4/internal/clients/telegram"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency/rates"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency/rates/cbr"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/expense"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/expense/limit"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/metrics"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
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

type storageFactory interface {
	CreateTelegramUserStorage() storage.TelegramUserStorage
	CreateExpenseStorage() storage.ExpenseStorage
	CreateLimitStorage() storage.ExpenseLimitStorage
	CreateCurrencyStorage() storage.CurrencyStorage
	CreateRatesStorage() storage.CurrencyRatesStorage
}

type client interface {
	RegisterController(model.Controller)
	ListenUpdates(ctx context.Context) error
}

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

			metricsServer := metrics.NewServer(uint16(metricsPort), logger)
			rater := rates.NewRater(cfg.Currency, storageFactory.CreateRatesStorage(), cbr.NewGateway(http.DefaultClient), logger)

			g.Go(func() error { return metricsServer.Run(ctx) })
			g.Go(func() error { return rater.Run(ctx) })

			finAssist := model.NewController(
				expense.NewExpenser(storageFactory.CreateExpenseStorage()),
				limit.NewLimiter(storageFactory.CreateLimitStorage()),
				currency.NewManager(cfg.Currency, storageFactory.CreateCurrencyStorage()),
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

func newStorageFactory(ctx context.Context, storageConfig config.StorageConfig, logger *zap.Logger) (storageFactory, error) {
	switch storageConfig.Driver {
	case config.InMemoryDriver:
		return inmemory.NewFactory(), nil

	case config.PostgreSQLDriver:
		return postgresql.NewFactory(ctx, storageConfig.Dsn, storageConfig.WaitTimeout, logger)
	}

	return nil, errors.New("unknown storage driver")
}

func newTelegramClient(token string, s storage.TelegramUserStorage, l *zap.Logger) (client, error) {
	return tgclient.NewClient(token, s, l)
}
