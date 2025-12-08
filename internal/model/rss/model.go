package rss

import (
	"encoding/xml"
	"time"
)

type Enclosure struct {
	URL  string `xml:"url,attr" json:"url"`
	Type string `xml:"type,attr" json:"type"`
}

type Item struct {
	Title         string     `xml:"title" json:"title"`
	PubDate       string     `xml:"pubDate" json:"pubDate"`
	PubTimeParsed *time.Time `xml:"-" json:"pubTimeParsed"`
	Category      []string   `xml:"category" json:"category"`
	Link          string     `xml:"link" json:"link"`
	AmpLink       string     `xml:"amplink" json:"ampLink"`
	Description   string     `xml:"description" json:"description"`
	Fulltext      string     `xml:"http://purl.org/rss/1.0/modules/content/ encoded" json:"fulltext"`
	Enclosure     Enclosure  `xml:"enclosure" json:"enclosure"`
	Guid          string     `xml:"guid" json:"guid"`
	Region        string     `xml:"region" json:"region"`
	Author        string     `xml:"author" json:"author"`
}

type Channel struct {
	Title       string `xml:"title" json:"title"`
	Link        string `xml:"link" json:"link"`
	Description string `xml:"description" json:"description"`
	Language    string `xml:"language" json:"language"`
	Items       []Item `xml:"item" json:"items"`
}

type Rss struct {
	Name    xml.Name `xml:"rss" json:"name"`
	Version string   `xml:"version,attr" json:"version"`
	Channel Channel  `xml:"channel" json:"channel"`
}
