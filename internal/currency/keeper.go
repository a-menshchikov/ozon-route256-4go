package currency

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
)

type Keeper struct {
	data            map[int64]string
	currencies      map[string]struct{}
	defaultCurrency string
}

func NewKeeper(cfg config.Currency) *Keeper {
	currencies := make(map[string]struct{})
	for _, currency := range cfg.Available {
		currencies[currency.Code] = struct{}{}
	}

	return &Keeper{
		data:            make(map[int64]string),
		currencies:      currencies,
		defaultCurrency: cfg.Base,
	}
}

func (k *Keeper) Set(userID int64, currency string) error {
	if _, ok := k.currencies[currency]; !ok {
		return ErrUnknownCurrency
	}

	k.data[userID] = currency

	return nil
}

func (k *Keeper) Get(userID int64) string {
	if currency, ok := k.data[userID]; ok {
		return currency
	}

	return k.defaultCurrency
}
