package implementation

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	model "gafarov/rss-reader/internal/model/kafka"
	"gafarov/rss-reader/internal/model/rss"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Kafka struct {
	writer   *kafka.Writer
	logger   *zap.Logger
	stopOnce sync.Once
}

func ping(addr ...string) error {
	for _, a := range addr {
		conn, err := kafka.Dial("tcp", a)
		if err != nil {
			return fmt.Errorf("kafka connect error: %w", err)
		}

		// проверка, что брокер отвечает
		_, err = conn.ReadPartitions()
		_ = conn.Close()

		if err != nil {
			return fmt.Errorf("kafka ping error: %w", err)
		}
	}
	return nil
}

func New(logger *zap.Logger, topic string, addr ...string) (*Kafka, error) {
	err := ping(addr...)
	if err != nil {

		if logger != nil {
			logger.Error("kafka ping error", zap.Error(err))
		}

		return nil, err
	}

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
		logger: logger,
	}, nil
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
		if k.logger != nil {
			k.logger.Error("json marshal error", zap.Error(err))
		}
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
		} else if k.logger != nil {
			k.logger.Error("kafka write error", zap.Error(err))
		}

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if k.logger != nil {
		k.logger.Info("kafka write ok")
	}

	return err
}
