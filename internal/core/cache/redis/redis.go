package redis

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func New(host, password string) (*RedisCache, error) {

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
		return nil, err
	}
	return data, nil
}

func (c *RedisCache) Set(key string, value []byte, expiration time.Duration) error {
	return c.client.Set(context.Background(), key, value, expiration).Err()
}
