package app

import (
	"context"
	"gafarov/rss-reader/internal/core/kafka"
	"gafarov/rss-reader/internal/core/reader"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type App struct {
	reader reader.IReader
	kafka  kafka.IKafka
	logger *zap.Logger
}

func New(reader reader.IReader, kafka kafka.IKafka, logger *zap.Logger) *App {
	return &App{
		reader: reader,
		kafka:  kafka,
		logger: logger,
	}
}

func (a *App) Run(url, code, name string, delay time.Duration, ctx context.Context) error {
	a.logger.Info("Starting app", zap.String("url", url), zap.String("code", code), zap.Duration("delay", delay))
	output := a.reader.Output()
	a.reader.StartParsing(url, name, delay, ctx)
	defer a.reader.Stop()

	channel, err := a.reader.GetChannel(url, ctx)
	if err != nil {
		a.logger.Error("failed to get channel", zap.Error(err))
		return err
	}

	isTesting := strings.ToLower(os.Getenv("TEST")) == "true"
	if isTesting {
		a.logger.Info("Running in testing mode")
	}

	for item := range output {
		err := a.kafka.Write(&item, channel, isTesting, code)
		if err != nil {
			a.logger.Error("failed to write item", zap.Error(err))
		}
	}
	return nil
}
