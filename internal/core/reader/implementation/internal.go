package implementation

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	LastReadGuidKey = "rss_reader:last_read_guid:"
)

func (r *RssReader) getLastReadGuid(name string) (string, error) {

	if r.cache == nil {
		if r.logger != nil {
			r.logger.Warn("cache is not initialized")
		}
		return "", fmt.Errorf("cache is not initialized")
	}

	data, err := r.cache.Get(LastReadGuidKey + name)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get last read guid", zap.Error(err))
		}
		return "", err
	}
	return string(data), nil
}

func (r *RssReader) saveLastReadGuid(guid, name string) error {

	if r.cache == nil {
		if r.logger != nil {
			r.logger.Warn("cache is not initialized")
		}
		return fmt.Errorf("cache is not initialized")
	}

	err := r.cache.Set(LastReadGuidKey+name, []byte(guid), 24*time.Hour)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to save last read guid", zap.Error(err))
		}
		return err
	} else if r.logger != nil {
		r.logger.Info("last read guid saved", zap.String("guid", guid))
	}

	return nil
}

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
