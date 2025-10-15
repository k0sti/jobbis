package utils

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Test set and get
	cache.Set("key1", "value1")

	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestCache_Expiry(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Set value with short TTL
	cache.Set("key1", "value1")

	// Should be found immediately
	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiry
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	cache := NewCache(1 * time.Hour)

	// Set with custom short TTL
	cache.SetWithTTL("key1", "value1", 100*time.Millisecond)

	// Should be found immediately
	_, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiry
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")

	if found1 || found2 {
		t.Error("Expected all keys to be cleared")
	}
}

func TestCache_MissingKey(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	_, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
}
