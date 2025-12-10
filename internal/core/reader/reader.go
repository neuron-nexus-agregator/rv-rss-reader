package reader

import (
	"context"
	"gafarov/rss-reader/internal/model/rss"
	"time"
)

type IReader interface {
	StartParsing(url string, delay time.Duration, ctx context.Context)
	ParseOnce(url string, ctx context.Context) []*rss.Item
	GetChannel(url string, ctx context.Context) (*rss.Channel, error)
	Output() <-chan rss.Item
}
