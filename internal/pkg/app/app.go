package app

import (
	"context"
	cache "gafarov/rss-reader/internal/core/cache/redis"
	kafka "gafarov/rss-reader/internal/core/kafka/implementation"
	reader "gafarov/rss-reader/internal/core/reader/implementation"
	endpoint "gafarov/rss-reader/internal/endpoint/app"
	"time"

	"go.uber.org/zap"
)

type RedisData struct {
	Host     string
	Password string
}

type KafkaData struct {
	Topic string
	Addr  []string
}

type App struct {
	endpoint *endpoint.App
	logger   *zap.Logger
}

func New(redisData RedisData, kafkaData KafkaData, logger *zap.Logger) (*App, error) {
	cache, err := cache.New(redisData.Host, redisData.Password, logger)
	if err != nil {
		logger.Error("failed to create cache", zap.Error(err))
		return nil, err
	}
	reader := reader.New(cache, logger)
	kafka, err := kafka.New(logger, kafkaData.Topic, kafkaData.Addr...)
	if err != nil {
		logger.Error("failed to create kafka", zap.Error(err))
		return nil, err
	}

	endpoint := endpoint.New(reader, kafka, logger)

	return &App{
		endpoint: endpoint,
		logger:   logger,
	}, nil
}

func (a *App) Run(url, name, code string, delay time.Duration, ctx context.Context) error {
	a.logger.Info("Starting app", zap.String("url", url), zap.String("name", name), zap.String("code", code), zap.Duration("delay", delay))
	return a.endpoint.Run(url, name, code, delay, ctx)
}
