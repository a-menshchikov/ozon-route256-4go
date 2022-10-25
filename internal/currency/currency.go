package currency

import (
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type manager struct {
	defaultCurrency string
	currencies      []currency
	storage         storage.CurrencyStorage
}

type currency struct {
	code string
	flag string
}

func NewManager(currencyCfg config.CurrencyConfig, s storage.CurrencyStorage) *manager {
	cc := make([]currency, len(currencyCfg.Available))
	for k, c := range currencyCfg.Available {
		cc[k] = currency{
			code: c.Code,
			flag: c.Flag,
		}
	}

	return &manager{
		defaultCurrency: currencyCfg.Base,
		currencies:      cc,
		storage:         s,
	}
}

func (m *manager) Get(user *types.User) (string, error) {
	if currency, found, err := m.storage.Get(user); err != nil {
		return "", errors.Wrap(err, "CurrencyStorage.Get")
	} else if found {
		return currency, nil
	}

	return m.defaultCurrency, nil
}

func (m *manager) Set(user *types.User, curr string) error {
	for _, c := range m.currencies {
		if c.code != curr {
			continue
		}

		return m.storage.Set(user, curr)
	}

	return errors.New("указана неизвестная валюта")
}

func (m *manager) ListCurrenciesCodesWithFlags() []string {
	list := make([]string, len(m.currencies))
	for k, c := range m.currencies {
		list[k] = c.code + " " + c.flag
	}

	return list
}
