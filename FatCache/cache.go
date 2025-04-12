package fatcache

import (
	"fatcache/lru"
	"fmt"
	"sync"
)

type Cache struct {
	cacheBytes int64
	lru        *lru.Cache
	mu         sync.Mutex
}

func (c *Cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru != nil {
		c.lru.Add(key, value)
	} else {
		c.lru = lru.New(c.cacheBytes, func(key string, value lru.Value) {
			fmt.Println("onEvict", key)
		})
		c.lru.Add(key, value)
	}
}

func (c *Cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}
	if value, ok := c.lru.Get(key); ok {
		return value.(ByteView), ok
	}
	return ByteView{}, false
}
