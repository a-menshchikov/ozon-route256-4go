package test

import (
	"context"
	"reflect"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

var (
	CtxInterface          = reflect.TypeOf((*context.Context)(nil)).Elem()
	UpdateConfigInterface = reflect.TypeOf((*tgbotapi.UpdateConfig)(nil)).Elem()

	TgUserID    = int64(123)
	User        = &([]types.User{types.User(TgUserID)}[0])
	Yesterday   = Today.Add(-time.Hour * 24)
	Today       = utils.TruncateToDate(time.Now())
	Tomorrow    = Today.Add(24 * time.Hour)
	SimpleError = errors.New("error")

	DefaultCurrencyCfg = config.CurrencyConfig{
		Base: "USD",
		Available: []config.Currency{
			{
				Code: "USD",
				Flag: "$",
			},
			{
				Code: "EUR",
				Flag: "¢",
			},
		},
		RefreshInterval: 100 * time.Millisecond,
	}
)
