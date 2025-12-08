package cache

import "time"

type LastRecord struct {
	Guid     string    `json:"guid"`
	ReadTime time.Time `json:"readTime"`
}
