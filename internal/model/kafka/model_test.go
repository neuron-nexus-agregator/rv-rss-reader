package kafka_test

import (
	"fmt"
	"gafarov/rss-reader/internal/model/kafka"
	"gafarov/rss-reader/internal/model/rss"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
)

func TestMessage_ConvertFromRSS(t *testing.T) {
	for i := range 30 {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			channel := &rss.Channel{
				Title:       gofakeit.Sentence(4),
				Link:        gofakeit.URL(),
				Description: gofakeit.Paragraph(3, 5, 15, " "),
				Language:    gofakeit.Language(),
			}

			testingChannel := &kafka.Channel{}
			result := testingChannel.ConvertFromRSS(channel)

			assert.Equal(t, channel.Title, testingChannel.Title, "Title должен совпасть")
			assert.Equal(t, channel.Link, testingChannel.Link, "Link должен совпасть")
			assert.Equal(t, channel.Description, testingChannel.Description, "Description должен совпасть")
			assert.Equal(t, channel.Language, testingChannel.Language, "Language должен совпасть")
			assert.Equal(t, testingChannel, result, "Объекты должны совпасть")
		})
	}
}
