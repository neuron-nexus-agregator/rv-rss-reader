package reader

import (
	"context"
	"gafarov/rss-reader/internal/model/rss"
	"time"
)

type IReader interface {
	StartParsing(url, name string, delay time.Duration, ctx context.Context) error
	ParseOnce(url string, ctx context.Context) ([]*rss.Item, error)
	GetChannel(url string, ctx context.Context) (*rss.Channel, error)
	Output() <-chan rss.Item
	Stop() error
}
