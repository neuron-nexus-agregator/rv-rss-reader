package kafka

import (
	"gafarov/rss-reader/internal/model/rss"
)

type IKafka interface {
	Write(item *rss.Item, channel *rss.Channel, isTesting bool, channelCode string) error
}
