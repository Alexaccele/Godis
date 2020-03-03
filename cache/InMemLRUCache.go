package cache

import (
	"Godis/cache/simplelru"
	"sync"
)

// InMemLRUCache is a thread-safe fixed size LRU cache.
type InMemLRUCache struct {
	lru  simplelru.LRUCache
	lock sync.RWMutex
}

// New creates an LRU of the given size.
func New() (*InMemLRUCache, error) {
	return NewWithEvict( nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvict(onEvicted func(key interface{}, value interface{})) (*InMemLRUCache, error) {
	lru, err := simplelru.NewLRU(simplelru.EvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &InMemLRUCache{
		lru: lru,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *InMemLRUCache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	c.lock.Unlock()
}

// Set adds a value to the cache. Returns true if an eviction occurred.
func (c *InMemLRUCache) Set(key, value interface{}) (evicted bool) {
	c.lock.Lock()
	evicted = c.lru.Set(key, value)
	c.lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (c *InMemLRUCache) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	value, ok = c.lru.Get(key)
	c.lock.Unlock()
	return value, ok
}

// Del removes the provided key from the cache.
func (c *InMemLRUCache) Del(key interface{}) (present bool) {
	c.lock.Lock()
	present = c.lru.Del(key)
	c.lock.Unlock()
	return
}

// RemoveOldest removes the oldest item from the cache.
func (c *InMemLRUCache) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	c.lock.Lock()
	key, value, ok = c.lru.RemoveOldest()
	c.lock.Unlock()
	return
}

// GetOldest returns the oldest entry
func (c *InMemLRUCache) GetOldest() (key interface{}, value interface{}, ok bool) {
	c.lock.Lock()
	key, value, ok = c.lru.GetOldest()
	c.lock.Unlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *InMemLRUCache) Keys() []interface{} {
	c.lock.RLock()
	keys := c.lru.Keys()
	c.lock.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *InMemLRUCache) Len() int {
	c.lock.RLock()
	length := c.lru.Len()
	c.lock.RUnlock()
	return length
}
