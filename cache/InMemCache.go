package cache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

/*
缓存的值结构
 */
type Value struct {
	Val     []byte
	Created time.Time
	TTL     time.Duration //生存时间
}

type entry struct {
	key string
	Value
}

type InMemCache struct {
	lock     sync.RWMutex
	evictList *list.List 		//LRU维护链表
	memCache map[string]*list.Element
	State
	memoryThreshold int64         //内存限制，0表示没有内存限制，单位字节Byte
	ExpireCycle     time.Duration //定期删除过期键值对周期
	strategy        ExpireStrategy
}

func (i *InMemCache) Get(k string) ([]byte, error) {
	//由于涉及到修改LRU链表，故加写锁，防止并发环境下，多协程修改节点位置导致内容丢失
	i.lock.Lock()
	defer i.lock.Unlock()
	if ent,exist := i.memCache[k]; exist{
		if ent.Value.(*entry) == nil{
			return nil,nil
		}
		val := ent.Value.(*entry).Value
		if val.TTL > 0 && val.Created.Add(val.TTL).Before(time.Now()){//设置了过期值
			//过期,惰性删除
			i.removeElement(ent)
			return nil,nil
		}
		i.evictList.MoveToFront(ent) //没被删除，则移至队列头部
		ent.Value.(*entry).Created = time.Now()//更新时间
		return val.Val,nil
	}
	return nil,nil
}

//SET操作只做存储，具体是否过期，应该由调用的上层接口定义
func (i *InMemCache) Set(k string, v Value) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	//检查空间
	for !i.checkMemory(k, v.Val) {
		i.strategy.MakeSpace(i)
	}
	//当存在时
	if ent, ok := i.memCache[k]; ok {
		i.evictList.MoveToFront(ent)
		ent.Value.(*entry).Value = v
		i.State.Del(k,ent.Value.(*entry).Value.Val)
		i.State.Add(k,v.Val)
		return nil
	}
	//新建
	ent := &entry{k, v}
	entry := i.evictList.PushFront(ent)
	i.memCache[k] = entry
	i.Add(k, v.Val)
	return nil
}
/*
不使用带过期时间的SET，Set只做放入，具体是否包含过期时间，由上层调用时决定
 */
//SET带过期时间的键值对
//func (i *InMemCache) SetWithTTL(k string, v []byte,ttl time.Duration) error {
//	i.lock.Lock()
//	defer i.lock.Unlock()
//	//检查空间
//	for !i.checkMemory(k, v) {
//		i.strategy.MakeSpace(i)
//	}
//	//当存在时
//	if ent, ok := i.memCache[k]; ok {
//		i.evictList.MoveToFront(ent)
//		i.State.Del(k,ent.Value.(*entry).Value.Val)
//		ent.Value.(*entry).Value = Value{v,time.Now(),ttl}//更新
//		i.Add(k,v)
//		return nil
//	}
//	//新建
//	ent := &entry{k, Value{v,time.Now(),ttl}}
//	entry := i.evictList.PushFront(ent)
//	i.memCache[k] = entry
//	i.Add(k, v)
//	return nil
//}

func (i *InMemCache) Del(k string) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	if ent, exist := i.memCache[k]; exist{
		i.removeElement(ent)
	}
	return nil
}

func (i *InMemCache) removeOldest() {
	ent := i.evictList.Back()
	if ent != nil {
		i.removeElement(ent)
	}
}

func (i *InMemCache) removeElement(e *list.Element) {
	i.evictList.Remove(e)
	kv := e.Value.(*entry)
	i.State.Del(kv.key, e.Value.(*entry).Val)
	delete(i.memCache, kv.key)
}

func (i *InMemCache) GetState() State {
	return i.State
}

func (i *InMemCache) checkMemory(k string, v []byte) bool {
	return i.memoryThreshold == 0 || i.State.Memory()+int64(len(k))+int64(len(v)) < i.memoryThreshold
}

func NewInMemCache(expireCycle time.Duration) *InMemCache {
	i := &InMemCache{
		lock:        sync.RWMutex{},
		memCache:    make(map[string]*list.Element),
		State:       State{},
		ExpireCycle: expireCycle,
	}
	if expireCycle > 0{
		go i.Expirer()
	}
	return i
}

func NewInMemCacheWithMemoryThreshold(memoryThreshold int64, expireCycle time.Duration) *InMemCache {
	i := &InMemCache{
		lock:            sync.RWMutex{},
		evictList:       list.New(),
		memCache:        make(map[string]*list.Element),
		State:           State{},
		memoryThreshold: memoryThreshold,
		ExpireCycle:     expireCycle,
		strategy:        &LRUAll{},
	}
	//if expireCycle > 0{
	//	go i.Expirer()
	//}
	return i
}

//定期删除过期键值对
func (cache *InMemCache) Expirer()  {
	for{
		time.Sleep(cache.ExpireCycle) //休眠
		cache.lock.RLock()
		for k,v := range cache.memCache{
			cache.lock.RUnlock()
			val := v.Value.(*entry).Value
			if val.TTL > 0 && val.Created.Add(val.TTL).Before(time.Now()){
				cache.Del(k)
			}
			cache.lock.RLock()
		}
		cache.lock.RUnlock()
	}
}


//返回从尾到头的顺序，用于恢复数据时，Set后能与原顺序保持一致
func (i *InMemCache) KeysAndValues() ([]string,[]Value) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	keys := make([]string,len(i.memCache))
	values := make([]Value,len(i.memCache))
	index := 0
	for ent := i.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[index] = ent.Value.(*entry).key
		values[index] = ent.Value.(*entry).Value
		index++
	}
	if index != len(i.memCache){
		log.Panicln("拷贝不完整")
	}
	return keys,values
}