package impl

import (
	"Godis/cache"
	"sync"
)

type inMemCache struct {
	lock     sync.RWMutex
	memCache map[string][]byte
	cache.State
	memoryThreshold uint64 //内存限制，0表示没有内存限制，单位字节Byte
}

func (i *inMemCache) Get(k string) ([]byte, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.memCache[k], nil
}

func (i *inMemCache) Set(k string, v []byte) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if !i.checkMemory(k, v) {
		//TODO 数据淘汰策略
	}
	bytes, exist := i.memCache[k]
	if exist {
		i.State.Del(k, bytes)
	}
	i.memCache[k] = v
	i.State.Add(k, v)
	return nil
}

func (i *inMemCache) Del(k string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	bytes, exist := i.memCache[k]
	if exist {
		i.State.Del(k, bytes)
		delete(i.memCache, k)
	}
	return nil
}

func (i *inMemCache) GetState() cache.State {
	return i.State
}

func (i *inMemCache) checkMemory(k string, v []byte) bool {
	return i.memoryThreshold == 0 || i.State.Memory()+uint64(int64(len(k))+int64(len(v))) < i.memoryThreshold
}

func NewInMemCache() *inMemCache {
	return &inMemCache{
		lock:     sync.RWMutex{},
		memCache: make(map[string][]byte),
		State:    cache.State{},
	}
}

func NewInMemCacheWithMemoryThreshold(memoryThreshold uint64) *inMemCache {
	return &inMemCache{
		lock:            sync.RWMutex{},
		memCache:        make(map[string][]byte),
		State:           cache.State{},
		memoryThreshold: memoryThreshold,
	}
}
