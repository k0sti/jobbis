package utils

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data   interface{}
	expiry time.Time
}

// Cache provides in-memory caching with TTL
type Cache struct {
	store      map[string]cacheEntry
	defaultTTL time.Duration
	mu         sync.RWMutex
}

// NewCache creates a new cache with the specified default TTL
func NewCache(defaultTTL time.Duration) *Cache {
	c := &Cache{
		store:      make(map[string]cacheEntry),
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.store[key]
	if !found {
		return nil, false
	}

	if time.Now().After(entry.expiry) {
		// Entry expired
		return nil, false
	}

	return entry.data, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = cacheEntry{
		data:   value,
		expiry: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]cacheEntry)
}

// cleanupLoop periodically removes expired entries
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.store {
		if now.After(entry.expiry) {
			delete(c.store, key)
		}
	}
}
