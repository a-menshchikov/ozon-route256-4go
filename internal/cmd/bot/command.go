package bot

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tgclient "gitlab.ozon.dev/almenschhikov/go-course-4/internal/clients/telegram"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency/rates"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency/rates/cbr"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/expense"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/expense/limit"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage/inmemory"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage/postgresql"
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
	var configPath string

	c := &cobra.Command{
		Use:           name,
		Short:         "Financial Assistant bot",
		Example:       name + " --config=.data/config.yaml",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,

		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.New(configPath)
			if err != nil {
				return errors.Wrap(err, "config init failed")
			}

			storageFactory, err := newStorageFactory(ctx, cfg.Storage)
			if err != nil {
				return errors.Wrap(err, "storage factory init failed")
			}

			rater := rates.NewRater(cfg.Currency, storageFactory.CreateRatesStorage(), cbr.NewGateway(http.DefaultClient))
			go rater.Run(ctx)

			finAssist := model.NewController(
				expense.NewExpenser(storageFactory.CreateExpenseStorage()),
				limit.NewLimiter(storageFactory.CreateLimitStorage()),
				currency.NewManager(cfg.Currency, storageFactory.CreateCurrencyStorage()),
				rater,
			)

			tgClient, err := newTelegramClient(cfg.Client.Telegram.Token, storageFactory.CreateTelegramUserStorage())
			if err != nil {
				return errors.Wrap(err, "telegram client init failed")
			}

			tgClient.RegisterController(finAssist)
			if err := tgClient.ListenUpdates(ctx); err != nil {
				return errors.Wrap(err, "telegram client run failed")
			}

			return nil
		},
	}

	c.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config path")

	return c
}

func newStorageFactory(ctx context.Context, storageConfig config.StorageConfig) (storageFactory, error) {
	switch storageConfig.Driver {
	case config.InMemoryDriver:
		return inmemory.NewFactory(), nil

	case config.PostgreSQLDriver:
		return postgresql.NewFactory(ctx, storageConfig.Dsn, storageConfig.WaitTimeout)
	}

	return nil, errors.New("unknown storage driver")
}

func newTelegramClient(token string, s storage.TelegramUserStorage) (client, error) {
	return tgclient.NewClient(token, s)
}
