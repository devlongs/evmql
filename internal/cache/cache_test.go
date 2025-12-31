package cache

import (
	"testing"
	"time"
)

func TestInMemoryCache_SetAndGet(t *testing.T) {
	cache := NewInMemoryCache(10, 5*time.Minute, 0)
	defer cache.Stop()

	cache.Set("key1", "value1", 0)

	value, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to find key1 in cache")
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := NewInMemoryCache(10, 100*time.Millisecond, 0)
	defer cache.Stop()

	cache.Set("key1", "value1", 100*time.Millisecond)

	// Should exist immediately
	_, ok := cache.Get("key1")
	if !ok {
		t.Error("Expected to find key1 in cache")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be expired")
	}
}

func TestInMemoryCache_MaxItems(t *testing.T) {
	cache := NewInMemoryCache(3, 5*time.Minute, 0)
	defer cache.Stop()

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	// Adding a 4th item should evict the oldest
	cache.Set("key4", "value4", 0)

	if cache.Size() != 3 {
		t.Errorf("Expected size to remain 3, got %d", cache.Size())
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache(10, 5*time.Minute, 0)
	defer cache.Stop()

	cache.Set("key1", "value1", 0)
	cache.Delete("key1")

	_, ok := cache.Get("key1")
	if ok {
		t.Error("Expected key1 to be deleted")
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := NewInMemoryCache(10, 5*time.Minute, 0)
	defer cache.Stop()

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func TestInMemoryCache_Cleanup(t *testing.T) {
	cache := NewInMemoryCache(10, 50*time.Millisecond, 100*time.Millisecond)
	defer cache.Stop()

	// Add items with short TTL
	cache.Set("key1", "value1", 50*time.Millisecond)
	cache.Set("key2", "value2", 50*time.Millisecond)

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Items should be cleaned up
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after cleanup, got %d", cache.Size())
	}
}

func TestInMemoryCache_Keys(t *testing.T) {
	cache := NewInMemoryCache(10, 5*time.Minute, 0)
	defer cache.Stop()

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	keys := cache.Keys()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestGenerateKey(t *testing.T) {
	key1 := GenerateKey("balance", "0x123", 1000)
	key2 := GenerateKey("balance", "0x123", 1000)
	key3 := GenerateKey("balance", "0x123", 2000)

	if key1 != key2 {
		t.Error("Same parameters should generate same key")
	}

	if key1 == key3 {
		t.Error("Different parameters should generate different keys")
	}
}

func TestNoOpCache(t *testing.T) {
	cache := NewNoOpCache()

	cache.Set("key1", "value1", 0)

	_, ok := cache.Get("key1")
	if ok {
		t.Error("NoOpCache should not cache anything")
	}

	if cache.Size() != 0 {
		t.Error("NoOpCache should always have size 0")
	}
}

func TestInMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewInMemoryCache(100, 5*time.Minute, 0)
	defer cache.Stop()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				cache.Set(string(rune(n*10+j)), j, 0)
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				cache.Get(string(rune(j)))
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Cache should still be functional
	cache.Set("test", "value", 0)
	_, ok := cache.Get("test")
	if !ok {
		t.Error("Cache should still work after concurrent access")
	}
}
