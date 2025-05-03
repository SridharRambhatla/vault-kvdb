package cache

import (
	"container/list"
	"sync"
	"time"

	"go-kvdb/pkg/types"
)

// LRUCache implements a thread-safe LRU cache
type LRUCache struct {
	capacity int64
	size     int64
	mu       sync.RWMutex
	cache    map[string]*list.Element
	list     *list.List
}

type entry struct {
	key    string
	value  *types.Context
	size   int64
	access time.Time
}

// NewLRUCache creates a new LRU cache with the given capacity in bytes
func NewLRUCache(capacity int64) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves a context from the cache
func (c *LRUCache) Get(key string) (*types.Context, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[key]; ok {
		c.list.MoveToFront(element)
		element.Value.(*entry).access = time.Now()
		return element.Value.(*entry).value, nil
	}
	return nil, nil
}

// Put adds or updates a context in the cache
func (c *LRUCache) Put(key string, value *types.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size (rough estimation)
	size := int64(len(key))
	for _, msg := range value.Messages {
		size += int64(len(msg.Role) + len(msg.Content))
	}

	// If key exists, update it
	if element, ok := c.cache[key]; ok {
		c.list.MoveToFront(element)
		oldSize := element.Value.(*entry).size
		element.Value.(*entry).value = value
		element.Value.(*entry).size = size
		element.Value.(*entry).access = time.Now()
		c.size = c.size - oldSize + size
		return
	}

	// If cache is full, remove least recently used items
	for c.size+size > c.capacity {
		element := c.list.Back()
		if element == nil {
			break
		}
		c.list.Remove(element)
		delete(c.cache, element.Value.(*entry).key)
		c.size -= element.Value.(*entry).size
	}

	// Add new entry
	entry := &entry{
		key:    key,
		value:  value,
		size:   size,
		access: time.Now(),
	}
	element := c.list.PushFront(entry)
	c.cache[key] = element
	c.size += size
}

// Remove removes a context from the cache
func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.cache[key]; ok {
		c.list.Remove(element)
		delete(c.cache, key)
		c.size -= element.Value.(*entry).size
	}
}

// List returns all contexts in the cache
func (c *LRUCache) List() []*types.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()

	contexts := make([]*types.Context, 0, len(c.cache))
	for _, element := range c.cache {
		contexts = append(contexts, element.Value.(*entry).value)
	}
	return contexts
}

// ListByTopic returns all contexts for a specific topic
func (c *LRUCache) ListByTopic(topic string) []*types.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()

	contexts := make([]*types.Context, 0)
	for _, element := range c.cache {
		ctx := element.Value.(*entry).value
		if ctx.Topic == topic {
			contexts = append(contexts, ctx)
		}
	}
	return contexts
}
