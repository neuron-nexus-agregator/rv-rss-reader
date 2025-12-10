package reader

import (
	"context"
	"gafarov/rss-reader/internal/model/rss"
	"time"
)

type IReader interface {
	StartParsing(url string, delay time.Duration, ctx context.Context)
	ParseOnce(url string, ctx context.Context) []*rss.Item
	Output() <-chan rss.Item
}
