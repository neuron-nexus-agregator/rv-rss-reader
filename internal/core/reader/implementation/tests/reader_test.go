package implementation_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	rss "gafarov/rss-reader/internal/core/reader/implementation"
)

var RSS_URL = "https://realnoevremya.ru/rss/yandex-dzen.xml"

func TestRssReader_IsStopped(t *testing.T) {
	r := rss.New()
	assert.False(t, r.IsStopped())

	err := r.Stop()
	assert.Nil(t, err)
	assert.True(t, r.IsStopped())
}

func TestRssReader_DoubleClose(t *testing.T) {
	r := rss.New()
	assert.False(t, r.IsStopped())

	err := r.Stop()
	assert.Nil(t, err)
	assert.True(t, r.IsStopped())

	err = r.Stop()
	assert.Equal(t, rss.ErrClosed, err)
}

func TestRssReader_StartParsing(t *testing.T) {
	r := rss.New()
	ctx := context.Background()

	err := r.StartParsing(RSS_URL, time.Second, ctx)
	assert.Nil(t, err)
	assert.False(t, r.IsStopped())

	err = r.Stop()
	assert.Nil(t, err)
	assert.True(t, r.IsStopped())
}

func TestRssReader_GetItems(t *testing.T) {
	r := rss.New()
	ctx := context.Background()

	items, err := r.ParseOnce(RSS_URL, ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, items)

	err = r.Stop()
	assert.Nil(t, err)
	assert.True(t, r.IsStopped())
}

func TestRssReader_GetItemContent(t *testing.T) {
	r := rss.New()
	ctx := context.Background()

	items, err := r.ParseOnce(RSS_URL, ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, items)

	item := items[0]
	assert.NotEmpty(t, item.Title)
	assert.NotNil(t, item.PubTimeParsed)
	assert.NotEmpty(t, item.Fulltext)
	assert.NotEmpty(t, item.Link)
	assert.NotEmpty(t, item.Description)

	err = r.Stop()
	assert.Nil(t, err)
	assert.True(t, r.IsStopped())
}

func TestRssReader_DoubleStart(t *testing.T) {
	r := rss.New()
	ctx := context.Background()

	err := r.StartParsing(RSS_URL, time.Second, ctx)
	assert.Nil(t, err)

	err = r.StartParsing(RSS_URL, time.Second, ctx)
	assert.Equal(t, rss.ErrAlreadyStarted, err)

	err = r.Stop()
	assert.Nil(t, err)
	assert.True(t, r.IsStopped())
}
