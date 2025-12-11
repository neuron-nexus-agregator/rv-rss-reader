package tests

import (
	"context"
	"gafarov/rss-reader/internal/core/kafka/implementation"
	"gafarov/rss-reader/internal/model/rss"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func createTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}

func deleteTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.DeleteTopics(topic)
}

func createMessage() (*rss.Item, *rss.Channel, bool, string) {
	item := &rss.Item{
		Title:       "Test Title",
		Description: "Test Description",
		Link:        "https://example.com",
	}

	channel := &rss.Channel{
		Title:       "Test Channel",
		Description: "Test Channel Description",
		Link:        "https://example.com",
	}

	return item, channel, true, "test_channel_code"
}

func TestKafka_Connection(t *testing.T) {
	_, err := implementation.New(nil, "test", os.Getenv("KAFKA_ADDR"))
	assert.NoError(t, err)
}

func TestKafka_Write(t *testing.T) {
	broker := os.Getenv("KAFKA_ADDR")
	topic := "test_topic_" + uuid.NewString()

	// 1. создаем уникальный топик
	err := createTopic(broker, topic)
	assert.NoError(t, err)

	// 2. создаем продюсер
	k, err := implementation.New(nil, topic, broker)
	assert.NoError(t, err)

	// 3. пишем сообщение
	m, c, b, s := createMessage()
	err = k.Write(m, c, b, s)
	assert.NoError(t, err)

	// 4. читаем сообщение, чтобы очистить топик
	conn, err := kafka.DialLeader(context.Background(), "tcp", broker, topic, 0)
	assert.NoError(t, err)
	_, err = conn.ReadMessage(1e6)
	assert.NoError(t, err)
	conn.Close()

	// 5. удаляем топик
	err = deleteTopic(broker, topic)
	assert.NoError(t, err)
}
