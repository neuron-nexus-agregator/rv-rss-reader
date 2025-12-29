package redis

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

func New(host, password string, logger *zap.Logger) (*RedisCache, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
	})

	err := client.Ping(context.Background()).Err()

	if err != nil {
		return nil, err
	}

	if os.Getenv("TEST") == "true" {
		_ = client.FlushDB(context.Background()).Err()
	}

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Get(key string) ([]byte, error) {
	data, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		if c.logger != nil {
			c.logger.Error("failed to get data from redis", zap.Error(err), zap.String("key", key))
		}
		return nil, err
	}

	return data, nil
}

func (c *RedisCache) Set(key string, value []byte, expiration time.Duration) error {
	err := c.client.Set(context.Background(), key, value, expiration).Err()

	if err != nil && c.logger != nil {
		c.logger.Error("failed to set data in redis", zap.Error(err), zap.String("key", key))
	} else if c.logger != nil {
		c.logger.Info("set in cache", zap.String("key", key), zap.String("value", string(value)))
	}

	return err
}
