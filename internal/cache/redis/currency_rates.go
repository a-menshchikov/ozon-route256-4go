package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"go.uber.org/zap"
)

type redisCurrencyRatesCache struct {
	keyPrefix string
	rdb       redis.Cmdable
	storage   storage.CurrencyRatesStorage
	logger    *zap.Logger
}

func NewCurrencyRatesCache(s storage.CurrencyRatesStorage, dsn string, l *zap.Logger) (*redisCurrencyRatesCache, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	return &redisCurrencyRatesCache{
		keyPrefix: "rate",
		rdb:       redis.NewClient(opts),
		storage:   s,
		logger:    l,
	}, nil
}

func (c *redisCurrencyRatesCache) Get(ctx context.Context, currency string, date time.Time) (int64, bool, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "redisCurrencyRatesCache.Get")
	defer span.Finish()

	var (
		key = c.cacheKey(currency, date)

		rate int64
		ok   bool
		err  error
	)

	cache, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		rate, err = strconv.ParseInt(cache, 10, 64)
		if err != nil {
			c.logger.Warn("cannot parse cache rate", zap.Error(err), zap.String("cache", cache), zap.String("key", key))
		} else {
			return rate, true, nil
		}
	} else if err != redis.Nil {
		c.logger.Warn("cannot get rates from redisReportCache", zap.Error(err))
	}

	rate, ok, err = c.storage.Get(ctx, currency, date)
	if ok && err == nil {
		if err := c.rdb.Set(ctx, key, rate, 0).Err(); err != nil {
			c.logger.Warn("cannot set rate cache", zap.Error(err), zap.Int64("rate", rate), zap.String("key", key))
		}
	}

	return rate, ok, err
}

func (c *redisCurrencyRatesCache) Add(ctx context.Context, currency string, date time.Time, rate int64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "redisCurrencyRatesCache.Add")
	defer span.Finish()

	key := c.cacheKey(currency, date)
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		c.logger.Warn("cannot delete rate from redisReportCache", zap.Error(err), zap.String("key", key))
	}

	return c.storage.Add(ctx, currency, date, rate)
}

func (c *redisCurrencyRatesCache) cacheKey(currency string, date time.Time) string {
	return fmt.Sprintf("%s_%s_%s", c.keyPrefix, currency, date)
}
