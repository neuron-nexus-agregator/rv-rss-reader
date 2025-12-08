package implementation_test

import (
	"fmt"
	"gafarov/rss-reader/internal/core/reader/implementation"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
)

func TestParseRSSDate(t *testing.T) {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		time.RFC3339Nano,
		time.UnixDate,
	}

	for i := range 30 {
		t.Run(fmt.Sprintf("Parse RSS Date %d", i+1), func(t *testing.T) {
			date := gofakeit.Date().UTC()
			for _, layout := range layouts {
				dateStr := date.Format(layout)
				dateParsed, err := implementation.ParseRSSDate(dateStr)

				assert.NoError(t, err)
				assert.NotNil(t, dateParsed)
				assert.Equal(t, dateParsed.Format(layout), dateStr, "Даты должны совпадать")
			}
		})
	}
}
