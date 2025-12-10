package app

import (
	"context"
	"gafarov/rss-reader/internal/core/cache"
	"gafarov/rss-reader/internal/core/kafka"
	"gafarov/rss-reader/internal/core/reader"
	"time"
)

type App struct {
	reader reader.IReader
	kafka  kafka.IKafka
}

func New(reader reader.IReader, kafka kafka.IKafka, cache cache.ICache) *App {
	return &App{
		reader: reader,
		kafka:  kafka,
	}
}

func (a *App) Run(url, code string, delay time.Duration, ctx context.Context) error {
	output := a.reader.Output()
	a.reader.StartParsing(url, delay, ctx)

	channel, err := a.reader.GetChannel(url, ctx)
	if err != nil {
		return err
	}

	for item := range output {
		err := a.kafka.Write(&item, channel, false, code)
		if err != nil {
			// TODO handle error
		}
	}
	return nil
}
