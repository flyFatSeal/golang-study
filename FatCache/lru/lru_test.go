package lru

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestNew(t *testing.T) {
	cache := New(1024, nil)
	assert.NotNil(t, cache)
	assert.Equal(t, int64(1024), cache.maxBytes)
	assert.Equal(t, int64(0), cache.nbytes)
	assert.NotNil(t, cache.ll)
	assert.NotNil(t, cache.cache)
}

func TestAddAndGet(t *testing.T) {
	cache := New(1024, nil)
	cache.Add("key1", String("value1"))
	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, String("value1"), value)

	// Test updating an existing key
	cache.Add("key1", String("newValue1"))
	value, ok = cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, String("newValue1"), value)
}

func TestRemoveOldest(t *testing.T) {
	cache := New(1024, nil)
	cache.Add("key1", String("value1"))
	cache.Add("key2", String("value2"))
	cache.Add("key3", String("value3"))

	// Manually remove the oldest item
	cache.RemoveOldest()
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	// Ensure other items are still present
	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, String("value2"), value)
}

func TestEviction(t *testing.T) {
	cache := New(10, nil)
	cache.Add("key1", String("1")) // 5 bytes
	cache.Add("key2", String("1")) // 5 bytes
	cache.Add("key3", String("1")) // 5 bytes, triggers eviction

	_, ok := cache.Get("key1")
	assert.False(t, ok) // key1 should be evicted

	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, String("1"), value)

	value, ok = cache.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, String("1"), value)
}

func TestOnEvict(t *testing.T) {
	evictedKeys := make([]string, 0)
	onEvict := func(key string, value Value) {
		evictedKeys = append(evictedKeys, key)
	}
	cache := New(10, onEvict)
	cache.Add("key1", String("12345")) // 5 bytes
	cache.Add("key2", String("12345")) // 5 bytes
	cache.Add("key3", String("12345")) // 5 bytes, triggers eviction

	assert.Equal(t, 1, len(evictedKeys))
	assert.Equal(t, "key1", evictedKeys[0])
}
func TestLen(t *testing.T) {
	cache := New(1000, nil)
	assert.Equal(t, 0, cache.Len())

	cache.Add("key1", String("12345"))
	assert.Equal(t, 1, cache.Len())

	cache.Add("key2", String("12345"))
	assert.Equal(t, 2, cache.Len())

	cache.RemoveOldest()
	assert.Equal(t, 1, cache.Len())
	cache.RemoveOldest()
	assert.Equal(t, 0, cache.Len())
	cache.RemoveOldest()
	assert.Equal(t, 0, cache.Len())
}

func TestStress(t *testing.T) {
	cache := New(1024*1024, nil) // 1 MB cache

	// Add a large number of items
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("key%d", i)
		value := String(fmt.Sprintf("value%d", i))
		cache.Add(key, value)
	}

	// Ensure the cache size does not exceed maxBytes
	assert.LessOrEqual(t, cache.nbytes, int64(1024*1024))

	// Check if recently added items are still accessible
	for i := 99990; i < 100000; i++ {
		key := fmt.Sprintf("key%d", i)
		value, ok := cache.Get(key)
		assert.True(t, ok)
		assert.Equal(t, String(fmt.Sprintf("value%d", i)), value)
	}

	// Check if oldest items are evicted
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		_, ok := cache.Get(key)
		assert.False(t, ok)
	}
}
