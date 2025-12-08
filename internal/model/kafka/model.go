package kafka

import (
	"gafarov/rss-reader/internal/model/rss"
)

type Message struct {
	NewsItem  *rss.Item `json:"newsItem"`
	Channel   *Channel  `json:"channel"`
	IsTesting bool      `json:"isTesting"`
}

type Channel struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Language    string `json:"language"`
}

func (c *Channel) ConvertFromRSS(channel *rss.Channel) *Channel {
	c.Title = channel.Title
	c.Link = channel.Link
	c.Description = channel.Description
	c.Language = channel.Language
	return c
}
