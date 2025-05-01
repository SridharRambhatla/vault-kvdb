package cache

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"go-kvdb/types"
)

var (
	ErrCacheFull = errors.New("cache is full")
	ErrNotFound  = errors.New("item not found in cache")
)

// CacheItem represents an item in the cache
type CacheItem struct {
	Key      string
	Value    *types.Context
	Size     int64
	LastUsed time.Time
	Element  *list.Element
}

// LRUCache implements a size-limited LRU cache
type LRUCache struct {
	mu       sync.RWMutex
	capacity int64
	used     int64
	list     *list.List
	items    map[string]*CacheItem
}

// NewLRUCache creates a new LRU cache with the given capacity in bytes
func NewLRUCache(capacity int64) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     list.New(),
		items:    make(map[string]*CacheItem),
	}
}

// Get retrieves an item from the cache
func (c *LRUCache) Get(key string) (*types.Context, error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, ErrNotFound
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update last used time
	item.LastUsed = time.Now()
	c.list.MoveToFront(item.Element)

	return item.Value, nil
}

// Put adds an item to the cache
func (c *LRUCache) Put(key string, value *types.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size of the new item
	size := int64(len(key)) + int64(len(value.Metadata.ID))

	// If item exists, remove it first
	if item, exists := c.items[key]; exists {
		c.list.Remove(item.Element)
		c.used -= item.Size
		delete(c.items, key)
	}

	// Evict items if necessary
	for c.used+size > c.capacity {
		if c.list.Len() == 0 {
			return ErrCacheFull
		}
		back := c.list.Back()
		item := back.Value.(*CacheItem)
		c.list.Remove(back)
		c.used -= item.Size
		delete(c.items, item.Key)
	}

	// Add new item
	element := c.list.PushFront(&CacheItem{
		Key:      key,
		Value:    value,
		Size:     size,
		LastUsed: time.Now(),
	})

	c.items[key] = &CacheItem{
		Key:      key,
		Value:    value,
		Size:     size,
		LastUsed: time.Now(),
		Element:  element,
	}
	c.used += size

	return nil
}

// Remove removes an item from the cache
func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		c.list.Remove(item.Element)
		c.used -= item.Size
		delete(c.items, key)
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.list.Init()
	c.items = make(map[string]*CacheItem)
	c.used = 0
}

// Size returns the current size of the cache in bytes
func (c *LRUCache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.used
}

// Count returns the number of items in the cache
func (c *LRUCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
