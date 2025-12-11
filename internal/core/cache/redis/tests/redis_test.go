package test

import (
	"gafarov/rss-reader/internal/core/cache/redis"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedis_Connect(t *testing.T) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PASSWORD")
	client, err := redis.New(host, port, nil)

	assert.Nil(t, err, "Ошибка подключения к Redis")
	assert.NotNil(t, client, "Клиент пустой")

	err = client.Close()
	assert.Nil(t, err, "Ошибка закрытия соединения с Redis")
}

func TestRedis_SetGetCorrect(t *testing.T) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PASSWORD")
	client, err := redis.New(host, port, nil)

	assert.Nil(t, err, "Ошибка подключения к Redis")
	assert.NotNil(t, client, "Клиент пустой")

	key := "test_key"
	value := "test_value"

	err = client.Set(key, []byte(value), 1*time.Minute)
	assert.Nil(t, err, "Ошибка записи в Redis")

	receivedValue, err := client.Get(key)
	assert.Nil(t, err, "Ошибка чтения из Redis")
	assert.Equal(t, value, string(receivedValue), "Значения не совпадают")

	err = client.Close()
	assert.Nil(t, err, "Ошибка закрытия соединения с Redis")
}

func TestRedis_SetGetNotCorrect(t *testing.T) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PASSWORD")
	client, err := redis.New(host, port, nil)

	assert.Nil(t, err, "Ошибка подключения к Redis")
	assert.NotNil(t, client, "Клиент пустой")

	key := "test_key"
	value := "test_value"

	err = client.Set(key, []byte(value), 1*time.Minute)
	assert.Nil(t, err, "Ошибка записи в Redis")

	receivedValue, err := client.Get(key + key)
	assert.Nil(t, err, "Ошибка чтения из Redis")
	assert.Empty(t, receivedValue)

	err = client.Close()
	assert.Nil(t, err, "Ошибка закрытия соединения с Redis")
}
