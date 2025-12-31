package cache

import (
	"testing"
	"time"
)

func TestInvalidateByPrefix(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	cache.Set("balance:0x123", "value1", 0)
	cache.Set("balance:0x456", "value2", 0)
	cache.Set("logs:0x789", "value3", 0)

	count := InvalidateByPrefix(cache, "balance:")

	if count != 2 {
		t.Errorf("Expected to invalidate 2 entries, got %d", count)
	}

	_, ok := cache.Get("balance:0x123")
	if ok {
		t.Error("Expected balance:0x123 to be invalidated")
	}

	_, ok = cache.Get("logs:0x789")
	if !ok {
		t.Error("Expected logs:0x789 to still exist")
	}
}

func TestInvalidateByAddress(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	// Add entries for different query types
	cache.Set("balance:0x123:some_hash", "value1", 0)
	cache.Set("logs:0x123:some_hash", "value2", 0)
	cache.Set("transactions:0x123:some_hash", "value3", 0)
	cache.Set("balance:0x456:some_hash", "value4", 0)

	count := InvalidateByAddress(cache, "0x123")

	if count != 3 {
		t.Errorf("Expected to invalidate 3 entries, got %d", count)
	}

	_, ok := cache.Get("balance:0x456:some_hash")
	if !ok {
		t.Error("Expected balance for 0x456 to still exist")
	}
}

func TestPatternInvalidator(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	invalidator := NewPatternInvalidator(cache)

	cache.Set("balance:0x123", "value1", 0)
	cache.Set("balance:0x456", "value2", 0)
	cache.Set("logs:0x789", "value3", 0)
	cache.Set("transactions:0xabc", "value4", 0)

	count := invalidator.InvalidateBalance()

	if count != 2 {
		t.Errorf("Expected to invalidate 2 balance entries, got %d", count)
	}

	_, ok := cache.Get("logs:0x789")
	if !ok {
		t.Error("Expected logs entry to still exist")
	}
}

func TestPatternInvalidator_InvalidateAll(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	invalidator := NewPatternInvalidator(cache)

	cache.Set("balance:0x123", "value1", 0)
	cache.Set("logs:0x456", "value2", 0)
	cache.Set("transactions:0x789", "value3", 0)

	invalidator.InvalidateAll()

	if cache.Size() != 0 {
		t.Errorf("Expected cache to be empty, got size %d", cache.Size())
	}
}

func TestTimeBasedInvalidation(t *testing.T) {
	strategy := &TimeBasedInvalidation{MaxAge: 1 * time.Hour}

	// Recent entry
	if strategy.ShouldInvalidate("key1", 30*time.Minute) {
		t.Error("Should not invalidate recent entries")
	}

	// Old entry
	if !strategy.ShouldInvalidate("key2", 2*time.Hour) {
		t.Error("Should invalidate old entries")
	}
}

func TestPrefixBasedInvalidation(t *testing.T) {
	strategy := &PrefixBasedInvalidation{Prefixes: []string{"balance:", "logs:"}}

	if !strategy.ShouldInvalidate("balance:0x123", 0) {
		t.Error("Should invalidate matching prefix")
	}

	if strategy.ShouldInvalidate("transactions:0x123", 0) {
		t.Error("Should not invalidate non-matching prefix")
	}
}

func TestInvalidateLogs(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	invalidator := NewPatternInvalidator(cache)

	cache.Set("balance:0x123", "value1", 0)
	cache.Set("logs:0x456", "value2", 0)
	cache.Set("logs:0x789", "value3", 0)

	count := invalidator.InvalidateLogs()

	if count != 2 {
		t.Errorf("Expected to invalidate 2 log entries, got %d", count)
	}

	_, ok := cache.Get("balance:0x123")
	if !ok {
		t.Error("Expected balance entry to still exist")
	}
}

func TestInvalidateTransactions(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	invalidator := NewPatternInvalidator(cache)

	cache.Set("transactions:0x123", "value1", 0)
	cache.Set("transactions:0x456", "value2", 0)
	cache.Set("balance:0x789", "value3", 0)

	count := invalidator.InvalidateTransactions()

	if count != 2 {
		t.Errorf("Expected to invalidate 2 transaction entries, got %d", count)
	}

	_, ok := cache.Get("balance:0x789")
	if !ok {
		t.Error("Expected balance entry to still exist")
	}
}
