package kafka

import (
	"gafarov/rss-reader/internal/model/rss"
)

type Message struct {
	NewsItem  rss.Item `json:"newsItem"`
	Channel   Channel  `json:"channel"`
	IsTesting bool     `json:"isTesting"`
}

type Channel struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Code        string `json:"codes"`
}

func (c *Channel) ConvertFromRSS(channel *rss.Channel, code string) *Channel {
	c.Title = channel.Title
	c.Link = channel.Link
	c.Description = channel.Description
	c.Language = channel.Language
	c.Code = code
	return c
}
