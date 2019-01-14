package plugins

import (
    "log"
    "sync"
    "time"
)

type Cache struct {
    lastFetchedAt *sync.Map
    resourceMap *sync.Map
    fetchInterval time.Duration
}

func NewCache(fetchInterval time.Duration) *Cache{
    cache := Cache{
        fetchInterval: fetchInterval,
        lastFetchedAt: new(sync.Map),
        resourceMap: new(sync.Map),
    }
    
    return &cache
}

func (c *Cache) ShouldFetch(key string) bool {
	v, ok := c.lastFetchedAt.Load(key)
	if !ok {
            log.Printf("[WARN] Not found %s in lastFetchedAt\n", key)
            return true
	}
	t, ok := v.(time.Time)
	if !ok {
            return true
	}
	return time.Since(t) > c.fetchInterval
}

func (c *Cache) UpdateLastFetchedAt(key string) {
    c.lastFetchedAt.Store(key, time.Now())
}

func (c *Cache) Store(key string, value interface{}) {
    c.resourceMap.Store(key, value)
}

func (c *Cache) Load(key string) interface{} {
    if ret, ok := c.resourceMap.Load(key); ok {
        return ret
    }
    return nil
}
