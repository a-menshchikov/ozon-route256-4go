package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type redisReportCache struct {
	keyPrefix string
	rdb       redis.Cmdable
	expenser  model.Expenser
	reporter  model.Reporter
	logger    *zap.Logger
}

func NewReportCache(e model.Expenser, r model.Reporter, dsn string, l *zap.Logger) (*redisReportCache, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	return &redisReportCache{
		keyPrefix: "report",
		rdb:       redis.NewClient(opts),
		expenser:  e,
		reporter:  r,
		logger:    l,
	}, nil
}

func (c *redisReportCache) AddExpense(ctx context.Context, user *types.User, date time.Time, amount int64, currency, category string) (err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "redisReportCache.AddExpense")
	defer span.Finish()

	defer func() {
		if err == nil {
			keyPattern := c.cacheKeyPattern(user, currency)
			if keys, err := c.rdb.Keys(ctx, keyPattern).Result(); err != nil {
				c.logger.Warn("cannot get cache keys", zap.Error(err), zap.String("pattern", keyPattern))
			} else if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				c.logger.Warn("cannot delete cache keys", zap.Error(err), zap.Strings("keys", keys))
			}
		}
	}()

	err = c.expenser.AddExpense(ctx, user, date, amount, currency, category)
	return
}

func (c *redisReportCache) GetReport(ctx context.Context, user *types.User, from time.Time, currency string) (map[string]int64, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "redisReportCache.GetReport")
	defer span.Finish()

	key := c.cacheKey(user, currency, from)

	cache, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var data map[string]int64
		if err = json.Unmarshal([]byte(cache), &data); err != nil {
			c.logger.Error("cannot unmarshal cached report", zap.Error(err), zap.String("key", key))
		} else {
			return data, nil
		}
	} else if err != redis.Nil {
		c.logger.Warn("cannot get report from cache", zap.Error(err), zap.String("key", key))
	}

	data, err := c.reporter.GetReport(ctx, user, from, currency)
	if err == nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			c.logger.Error("cannot marshal report", zap.Error(err))
		} else if err := c.rdb.Set(ctx, key, bytes, 0).Err(); err != nil {
			c.logger.Warn("cannot set report cache", zap.Error(err), zap.String("key", key), zap.ByteString("cache", bytes))
		}
	}

	return data, err
}

func (c *redisReportCache) cacheKey(user *types.User, currency string, date time.Time) string {
	return fmt.Sprintf("%s_%d_%s_%s", c.keyPrefix, int64(*user), currency, date)
}

func (c *redisReportCache) cacheKeyPattern(user *types.User, currency string) string {
	return fmt.Sprintf("%s_%d_%s_*", c.keyPrefix, int64(*user), currency)
}
