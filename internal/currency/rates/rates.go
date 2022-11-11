package rates

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/cache"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"go.uber.org/zap"
)

const (
	_cachePrefix = "rate"
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
	cache   cache.Cache
	gateway gateway
	logger  *zap.Logger
}

func NewRater(currencyCfg config.CurrencyConfig, s storage.CurrencyRatesStorage, cache cache.Cache, g gateway, l *zap.Logger) *rater {
	return &rater{
		refreshInterval: currencyCfg.RefreshInterval,
		baseCurrency:    currencyCfg.Base,

		storage: s,
		cache:   cache,
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

func (r *rater) Ready() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.ready
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
		rate   int64
		cached string
		err    error
		ok     bool
	)

	cacheKey := fmt.Sprintf("%s_%s_%s", _cachePrefix, from, date)
	if cached, ok = r.cache.Get(cacheKey); ok {
		r.logger.Debug("cached rate (from)", zap.String("rate", cached))
		rate, err = strconv.ParseInt(cached, 10, 64)
	} else {
		r.logger.Debug("there is no rate in cache", zap.String("key", cacheKey))
	}

	if !ok || err != nil {
		rate, ok, err = r.storage.Get(ctx, from, date)
		if err != nil {
			return 0, errors.Wrap(err, "CurrencyRatesStorage.Get")
		} else if !ok {
			return 0, errCannotExchange
		}

		if rate != 0 {
			if err := r.cache.Set(cacheKey, rate); err != nil {
				r.logger.Warn("cannot set rate to cache", zap.Error(err), zap.String("key", cacheKey), zap.Int64("rate", rate))
			}
		}
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
		cacheKey := fmt.Sprintf("%s_%s_%s", _cachePrefix, curr, date)
		if err := r.storage.Add(ctx, curr, date, rate); err != nil {
			r.logger.Error("CurrencyRatesStorage.Add failed", zap.Error(err))
		} else if err := r.cache.DeleteByPattern(cacheKey); err != nil {
			r.logger.Warn("cannot rate from cache", zap.Error(err), zap.String("key", cacheKey))
		}
	}
	r.ready = true
}
