package implementation

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
		BatchTimeout: 5 * time.Second,
		BatchSize:    5,
		BatchBytes:   1e6,
		Async:        false,
	})

	return &Kafka{
		writer: writer,
	}
}

func (k *Kafka) Close() {
	k.stopOnce.Do(func() {
		_ = k.writer.Close()
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

	for i := range 3 {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err = k.writer.WriteMessages(ctx, kafka.Message{
			Value: data,
		})
		cancel()

		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return err
}
