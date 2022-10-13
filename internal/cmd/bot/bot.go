package bot

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/clients/tg"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/currency/exchanger"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/expense"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
)

func NewCommand(name, version string) *cobra.Command {
	var configPath string

	c := &cobra.Command{
		Use:           name,
		Short:         "Financial Telegram bot",
		Example:       name + " --config=data/config.yaml",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,

		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.New(configPath)
			if err != nil {
				return errors.Wrap(err, "config init failed")
			}

			tgClient, err := tg.New(cfg.Token)
			if err != nil {
				return errors.Wrap(err, "tg client init failed")
			}

			cbrExchanger := exchanger.NewCbrExchanger(cfg.Currency)
			go cbrExchanger.Run(ctx)

			bot := model.NewBot(
				tgClient,
				expense.NewInmemoryStorage(),
				cbrExchanger,
				currency.NewKeeper(cfg.Currency),
			)

			tgClient.ListenUpdates(ctx, bot)

			return nil
		},
	}

	c.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config path")

	return c
}
