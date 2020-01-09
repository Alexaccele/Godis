package impl

import (
	"Godis/cache"
	"sync"
)

type inMemCache struct {
	lock sync.RWMutex
	memCache map[string][]byte
	cache.State
}

func (i *inMemCache) Get(k string) ([]byte, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.memCache[k],nil
}

func (i *inMemCache) Set(k string, v []byte) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	bytes,exist := i.memCache[k]
	if exist{
		i.State.Del(k,bytes)
	}
	i.memCache[k] = v
	i.State.Add(k,v)
	return nil
}

func (i *inMemCache) Del(k string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	bytes,exist := i.memCache[k]
	if exist{
		i.State.Del(k,bytes)
		delete(i.memCache,k)
	}
	return nil
}

func (i *inMemCache) GetState() cache.State {
	return i.State
}

func NewInMemCache() *inMemCache {
	return &inMemCache{
		lock:     sync.RWMutex{},
		memCache: make(map[string][]byte),
		State:    cache.State{},
	}
}