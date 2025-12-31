package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// Cache interface defines the caching operations
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
	Keys() []string
}

// cacheItem represents a single cached item with expiration
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// InMemoryCache implements an in-memory cache with TTL support
type InMemoryCache struct {
	items        map[string]*cacheItem
	mu           sync.RWMutex
	maxItems     int
	defaultTTL   time.Duration
	cleanupEvery time.Duration
	stopCleanup  chan struct{}
}

// NewInMemoryCache creates a new in-memory cache with the given configuration
func NewInMemoryCache(maxItems int, defaultTTL, cleanupEvery time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		items:        make(map[string]*cacheItem),
		maxItems:     maxItems,
		defaultTTL:   defaultTTL,
		cleanupEvery: cleanupEvery,
		stopCleanup:  make(chan struct{}),
	}

	// Start cleanup goroutine if cleanup interval is set
	if cleanupEvery > 0 {
		go cache.cleanupLoop()
	}

	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Set adds or updates a value in the cache
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Check if we need to evict items
	if len(c.items) >= c.maxItems {
		c.evictOldest()
	}

	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}

// Size returns the number of items in the cache
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Keys returns all cache keys
func (c *InMemoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}

// Stop stops the cleanup goroutine
func (c *InMemoryCache) Stop() {
	close(c.stopCleanup)
}

// cleanupLoop periodically removes expired items
func (c *InMemoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes expired items
func (c *InMemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}

// evictOldest removes the oldest item from the cache
func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiration
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// GenerateKey creates a cache key from the given parameters
func GenerateKey(prefix string, params ...interface{}) string {
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	return prefix + ":" + hex.EncodeToString(hash[:])
}

// NoOpCache implements a cache that doesn't actually cache anything
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (n *NoOpCache) Get(key string) (interface{}, bool)                   { return nil, false }
func (n *NoOpCache) Set(key string, value interface{}, ttl time.Duration) {}
func (n *NoOpCache) Delete(key string)                                    {}
func (n *NoOpCache) Clear()                                               {}
func (n *NoOpCache) Size() int                                            { return 0 }
func (n *NoOpCache) Keys() []string                                       { return []string{} }
