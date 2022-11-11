package redis

import (
	"context"

	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
)

type cache struct {
	rdb    redis.Cmdable
	ctx    context.Context
	logger *zap.Logger
}

func NewCache(ctx context.Context, dsn string, l *zap.Logger) (*cache, error) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	return &cache{
		ctx:    ctx,
		rdb:    redis.NewClient(opts),
		logger: l,
	}, nil
}

func (c *cache) Get(key string) (string, bool) {
	data, err := c.rdb.Get(c.ctx, key).Result()
	if err != nil {
		return "", false
	}

	return data, true
}

func (c *cache) Set(key string, value any) error {
	return c.rdb.Set(c.ctx, key, value, 0).Err()
}

func (c *cache) DeleteByPattern(pattern string) error {
	keys, err := c.rdb.Keys(c.ctx, pattern).Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := c.rdb.Del(c.ctx, key).Err(); err != nil {
			return err
		}
	}

	return nil
}
