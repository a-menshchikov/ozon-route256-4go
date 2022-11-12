package currency

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type currencyManager struct {
	defaultCurrency string
	currencies      []currency
	storage         storage.CurrencyStorage
}

type currency struct {
	code string
	flag string
}

func NewCurrencyManager(currencyCfg config.CurrencyConfig, s storage.CurrencyStorage) *currencyManager {
	cc := make([]currency, len(currencyCfg.Available))
	for k, c := range currencyCfg.Available {
		cc[k] = currency{
			code: c.Code,
			flag: c.Flag,
		}
	}

	return &currencyManager{
		defaultCurrency: currencyCfg.Base,
		currencies:      cc,
		storage:         s,
	}
}

func (m *currencyManager) Get(ctx context.Context, user *types.User) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "currencyManager.Get", opentracing.Tags{
		"user": *user,
	})
	defer span.Finish()

	if currency, found, err := m.storage.Get(ctx, user); err != nil {
		return "", errors.Wrap(err, "CurrencyStorage.Get")
	} else if found {
		return currency, nil
	}

	return m.defaultCurrency, nil
}

func (m *currencyManager) Set(ctx context.Context, user *types.User, curr string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "currencyManager.Set", opentracing.Tags{
		"user":     *user,
		"currency": curr,
	})
	defer span.Finish()

	for _, c := range m.currencies {
		if c.code != curr {
			continue
		}

		return m.storage.Set(ctx, user, curr)
	}

	return errors.New("указана неизвестная валюта")
}

func (m *currencyManager) ListCurrenciesCodesWithFlags() []string {
	list := make([]string, len(m.currencies))
	for k, c := range m.currencies {
		list[k] = c.code + " " + c.flag
	}

	return list
}
