package reader

import (
	"gafarov/rss-reader/internal/model/rss"
	"time"
)

type IReader interface {
	StartParsing(url string, delay time.Duration)
	ParseOnce(url string) []*rss.Item
}
