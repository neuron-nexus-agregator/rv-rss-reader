package implementation

import (
	"fmt"
	"time"
)

func ParseRSSDate(s string) (*time.Time, error) {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		time.RFC3339Nano,
		time.UnixDate,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("cannot parse date: %s", s)
}
