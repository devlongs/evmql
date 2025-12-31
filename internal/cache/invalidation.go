package cache

import (
	"strings"
	"time"
)

// InvalidationStrategy defines how cache entries should be invalidated
type InvalidationStrategy interface {
	ShouldInvalidate(key string, age time.Duration) bool
}

// TimeBasedInvalidation invalidates entries older than a certain age
type TimeBasedInvalidation struct {
	MaxAge time.Duration
}

func (t *TimeBasedInvalidation) ShouldInvalidate(key string, age time.Duration) bool {
	return age > t.MaxAge
}

// PrefixBasedInvalidation invalidates entries matching certain prefixes
type PrefixBasedInvalidation struct {
	Prefixes []string
}

func (p *PrefixBasedInvalidation) ShouldInvalidate(key string, age time.Duration) bool {
	for _, prefix := range p.Prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// InvalidateByPrefix removes all cache entries with the given prefix
func InvalidateByPrefix(c Cache, prefix string) int {
	count := 0
	keys := c.Keys()

	for _, key := range keys {
		if strings.HasPrefix(key, prefix) {
			c.Delete(key)
			count++
		}
	}

	return count
}

// InvalidateByAddress removes all cache entries for a specific address
func InvalidateByAddress(c Cache, address string) int {
	return InvalidateByPrefix(c, "balance:"+address) +
		InvalidateByPrefix(c, "logs:"+address) +
		InvalidateByPrefix(c, "transactions:"+address)
}

// InvalidateOlderThan removes all cache entries older than the specified duration
// Note: This requires the cache implementation to track creation/access times
func InvalidateOlderThan(c Cache, maxAge time.Duration) int {
	// This is a simplified version - a full implementation would need
	// the cache to track item ages
	count := 0

	// For now, this would need to be implemented by the specific cache type
	// The InMemoryCache already handles this via its cleanup mechanism

	return count
}

// InvalidatePattern removes cache entries matching a specific pattern
type PatternInvalidator struct {
	cache Cache
}

// NewPatternInvalidator creates a new pattern-based invalidator
func NewPatternInvalidator(c Cache) *PatternInvalidator {
	return &PatternInvalidator{cache: c}
}

// InvalidateBalance invalidates all balance cache entries
func (p *PatternInvalidator) InvalidateBalance() int {
	return InvalidateByPrefix(p.cache, "balance:")
}

// InvalidateLogs invalidates all log cache entries
func (p *PatternInvalidator) InvalidateLogs() int {
	return InvalidateByPrefix(p.cache, "logs:")
}

// InvalidateTransactions invalidates all transaction cache entries
func (p *PatternInvalidator) InvalidateTransactions() int {
	return InvalidateByPrefix(p.cache, "transactions:")
}

// InvalidateAll clears the entire cache
func (p *PatternInvalidator) InvalidateAll() {
	p.cache.Clear()
}
