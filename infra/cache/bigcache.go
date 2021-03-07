package cache

import (
	"time"

	"github.com/allegro/bigcache/v3"
)

// NewLocalCache is the factory of BigCache
func NewLocalCache() (*bigcache.BigCache, error) {
	cacheConfig := bigcache.DefaultConfig(10 * time.Minute)
	cache, err := bigcache.NewBigCache(cacheConfig)
	if err != nil {
		return nil, err
	}
	return cache, nil
}
