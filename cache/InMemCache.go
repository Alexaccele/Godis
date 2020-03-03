package cache

import (
	"sync"
	"time"
)

/*
缓存的值结构
 */
type value struct {
	Val     []byte
	Created time.Time
	TTL     time.Duration //生存时间
}

type InMemCache struct {
	lock     sync.RWMutex
	memCache map[string]value
	State
	memoryThreshold int64        //内存限制，0表示没有内存限制，单位字节Byte
	expireCycle     time.Duration //生存时间
	strategy        ExpireStrategy
}

func (i *InMemCache) Get(k string) ([]byte, error) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	v,exist := i.memCache[k]
	if !exist{//不存在
		return nil,nil
	}
	if v.TTL > 0 && v.Created.Add(v.TTL).Before(time.Now()){ //设置了过期值
		//过期,惰性删除
		i.Del(k)
		return nil,nil
	}
	return v.Val, nil
}

//默认SET不过期
func (i *InMemCache) Set(k string, v []byte) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if !i.checkMemory(k, v) {
		i.strategy.MakeSpace(i)
	}
	bytes, exist := i.memCache[k]
	if exist {
		i.State.Del(k, bytes.Val)
	}
	i.memCache[k] = value{v, time.Now(),-1}
	i.Add(k, v)
	return nil
}
//SET带过期时间的键值对
func (i *InMemCache) SetWithTTL(k string, v []byte,ttl time.Duration) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if !i.checkMemory(k, v) {
		i.strategy.MakeSpace(i)
	}
	bytes, exist := i.memCache[k]
	if exist {
		i.State.Del(k, bytes.Val)
	}
	i.memCache[k] = value{v, time.Now(),ttl}
	i.Add(k, v)
	return nil
}

func (i *InMemCache) Del(k string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	bytes, exist := i.memCache[k]
	if exist {
		i.State.Del(k, bytes.Val)
		delete(i.memCache, k)
	}
	return nil
}

func (i *InMemCache) GetState() State {
	return i.State
}

func (i *InMemCache) checkMemory(k string, v []byte) bool {
	return i.memoryThreshold == 0 || i.State.Memory()+int64(len(k))+int64(len(v)) < i.memoryThreshold
}

func NewInMemCache(ttl time.Duration) *InMemCache {
	i := &InMemCache{
		lock:        sync.RWMutex{},
		memCache:    make(map[string]value),
		State:       State{},
		expireCycle: ttl,
	}
	if ttl > 0{
		go i.expirer()
	}
	return i
}

func NewInMemCacheWithMemoryThreshold(memoryThreshold int64, expireCycle time.Duration) *InMemCache {
	i := &InMemCache{
		lock:            sync.RWMutex{},
		memCache:        make(map[string]value),
		State:           State{},
		memoryThreshold: memoryThreshold,
		expireCycle:     expireCycle,
		strategy:        &RandAll{},
	}
	if expireCycle > 0{
		go i.expirer()
	}
	return i
}

//定期删除过期键值对
func (cache *InMemCache) expirer()  {
	for{
		time.Sleep(cache.expireCycle) //休眠
		cache.lock.RLock()
		for k,v := range cache.memCache{
			cache.lock.RUnlock()
			if v.TTL > 0 && v.Created.Add(v.TTL).Before(time.Now()){
				cache.Del(k)
			}
			cache.lock.RUnlock()
		}
	}
}