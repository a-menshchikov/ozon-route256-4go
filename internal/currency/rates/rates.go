package rates

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
)

var (
	errCannotExchange = errors.New("не удалось выполнить конвертацию валюты")
)

type gateway interface {
	FetchRates(ctx context.Context) (map[string]int64, time.Time, error)
}

type rater struct {
	mu sync.RWMutex

	ready           bool
	refreshInterval time.Duration
	baseCurrency    string

	storage storage.CurrencyRatesStorage
	gateway gateway
}

func NewRater(currencyCfg config.CurrencyConfig, s storage.CurrencyRatesStorage, g gateway) *rater {
	return &rater{
		refreshInterval: currencyCfg.RefreshInterval,
		baseCurrency:    currencyCfg.Base,

		storage: s,
		gateway: g,
	}
}

func (r *rater) Run(ctx context.Context) {
	r.refreshRates(ctx)
	ticker := time.NewTicker(r.refreshInterval)

	select {
	case <-ctx.Done():
		ticker.Stop()
		return

	case <-ticker.C:
		r.refreshRates(ctx)
	}
}

func (r *rater) Ready() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.ready
}

func (r *rater) Exchange(value int64, from, to string, date time.Time) (int64, error) {
	if from == to {
		return value, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var (
		fromRate = int64(10000)
		toRate   = int64(10000)

		err error
		ok  bool
	)

	if from != r.baseCurrency {
		fromRate, ok, err = r.storage.Get(from, date)
		if err != nil {
			return 0, errors.Wrap(err, "CurrencyRatesStorage.Get (from)")
		} else if !ok {
			return 0, errCannotExchange
		}

	}

	if to != r.baseCurrency {
		toRate, ok, err = r.storage.Get(to, date)
		if err != nil {
			return 0, errors.Wrap(err, "CurrencyRatesStorage.Get (to)")
		} else if !ok {
			return 0, errCannotExchange
		}
	}

	return value * fromRate / toRate, nil
}

func (r *rater) refreshRates(ctx context.Context) {
	rates, date, err := r.gateway.FetchRates(ctx)
	if err != nil {
		log.Println("rates refresh failed:", err.Error())
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.ready = false
	for curr, rate := range rates {
		if err := r.storage.Add(curr, date, rate); err != nil {
			log.Println("CurrencyRatesStorage.Add failed:", err.Error())
		}
	}
	r.ready = true
}
