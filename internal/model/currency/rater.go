package currency

import (
	"context"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"go.uber.org/zap"
)

var (
	errCannotExchange = errors.New("не удалось выполнить конвертацию валюты")
)

type gateway interface {
	FetchRates(ctx context.Context) (map[string]int64, time.Time, error)
}

type rater struct {
	mu *sync.RWMutex

	ready           bool
	refreshInterval time.Duration
	baseCurrency    string

	storage storage.CurrencyRatesStorage
	gateway gateway
	logger  *zap.Logger
}

func NewRater(currencyCfg config.CurrencyConfig, s storage.CurrencyRatesStorage, g gateway, l *zap.Logger) *rater {
	return &rater{
		mu: new(sync.RWMutex),

		refreshInterval: currencyCfg.RefreshInterval,
		baseCurrency:    currencyCfg.Base,

		storage: s,
		gateway: g,
		logger:  l,
	}
}

func (r *rater) Run(ctx context.Context) error {
	r.refreshRates(ctx)
	ticker := time.NewTicker(r.refreshInterval)

	select {
	case <-ctx.Done():
		ticker.Stop()
		break

	case <-ticker.C:
		r.refreshRates(ctx)
	}

	return nil
}

func (r *rater) TryAcquireExchange() bool {
	return r.mu.TryRLock()
}

func (r *rater) ReleaseExchange() {
	r.mu.RUnlock()
}

func (r *rater) Exchange(ctx context.Context, value int64, from, to string, date time.Time) (int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "rater.Exchange", opentracing.Tags{
		"value": value,
		"from":  from,
		"to":    to,
		"date":  date,
	})
	defer span.Finish()

	if from == to {
		return value, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var (
		fromRate = int64(10000)
		toRate   = int64(10000)

		err error
	)

	if from != r.baseCurrency {
		fromRate, err = r.getRate(ctx, from, date)
		if err != nil {
			return 0, err
		}
	}

	if to != r.baseCurrency {
		toRate, err = r.getRate(ctx, to, date)
		if err != nil {
			return 0, err
		}
	}

	return value * fromRate / toRate, nil
}

func (r *rater) getRate(ctx context.Context, from string, date time.Time) (int64, error) {
	var (
		rate int64
		err  error
		ok   bool
	)

	rate, ok, err = r.storage.Get(ctx, from, date)
	if err != nil {
		return 0, errors.Wrap(err, "CurrencyRatesStorage.Get")
	} else if !ok {
		return 0, errCannotExchange
	}

	return rate, nil
}

func (r *rater) refreshRates(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "rater.refreshRates")
	defer span.Finish()

	rates, date, err := r.gateway.FetchRates(ctx)
	if err != nil {
		r.logger.Warn("rates refresh failed", zap.Error(err))
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.ready = false
	for curr, rate := range rates {
		if err := r.storage.Add(ctx, curr, date, rate); err != nil {
			r.logger.Error("CurrencyRatesStorage.Add failed", zap.Error(err))
		}
	}
	r.ready = true
}
