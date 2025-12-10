package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	model "gafarov/rss-reader/internal/model/kafka"
	"gafarov/rss-reader/internal/model/rss"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	writer   *kafka.Writer
	stopOnce sync.Once
}

func New(topic string, addr ...string) *Kafka {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      addr,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireAll),
		Async:        false,
	})

	return &Kafka{
		writer: writer,
	}
}

func (k *Kafka) Close() {
	k.stopOnce.Do(func() {
		k.writer.Close()
	})
}

func (k *Kafka) Write(item *rss.Item, channel *rss.Channel, isTesting bool, channelCode string) error {
	kafkaItem := model.Message{
		NewsItem:  *item,
		IsTesting: isTesting,
	}
	kafkaChannel := &model.Channel{}
	kafkaChannel.ConvertFromRSS(channel, channelCode)
	kafkaItem.Channel = *kafkaChannel
	data, err := json.Marshal(kafkaItem)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err = k.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})

	if err != nil {
		return err
	}
	return nil
}
