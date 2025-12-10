package kafka

import (
	"gafarov/rss-reader/internal/model/rss"
)

type Kafka interface {
	Write(item *rss.Item, channel *rss.Channel, isTesting bool, channelCode string) error
}
