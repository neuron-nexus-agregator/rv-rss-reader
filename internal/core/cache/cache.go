package cache

import "time"

type ICache interface {
	Get(key string) ([]byte, error)
	Set(key string, value any, ttl time.Duration) error
}
