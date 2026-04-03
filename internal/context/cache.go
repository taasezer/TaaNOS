package context

import (
	"sync"
	"time"
)

// Cache provides a time-limited cache for system context data.
// Avoids redundant system calls within a single pipeline run.
type Cache struct {
	mu       sync.RWMutex
	data     map[string]cacheEntry
	ttl      time.Duration
}

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewCache creates a context cache with the given TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]cacheEntry),
		ttl:  ttl,
	}
}

// Get retrieves a cached value. Returns nil, false if not found or expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all cached entries.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
}
