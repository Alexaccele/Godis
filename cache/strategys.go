package cache

/*
过期策略
*/
type ExpireStrategy interface {
	MakeSpace(cache *InMemCache)
}

/*
从所有键值对中随机删除
*/
type RandAll struct{}

func (r *RandAll) MakeSpace(cache *InMemCache) {
	toDelCount := cache.Count / 10 //默认淘汰10%的数据
	if toDelCount <= 0 {
		toDelCount = 1
	}
	count := 0
	for k, _ := range cache.memCache { //golang 实现的for-range map是随机化的，故直接遍历
		if count < toDelCount {
			if ent, exist := cache.memCache[k]; exist {
				cache.removeElement(ent)
				count++
			}
		}
		break
	}
}

/*
从所有带过期时间的键值对中随机删除
*/
type RandVolatile struct{}

func (r *RandVolatile) MakeSpace(cache *InMemCache) {
	toDelCount := cache.Count / 10
	if toDelCount <= 0 {
		toDelCount = 1
	}
	count := 0
	for k, v := range cache.memCache {
		if v.Value.(*entry).TTL > 0 {
			if count < toDelCount {
				if ent, exist := cache.memCache[k]; exist {
					cache.removeElement(ent)
					count++
				}
			} else {
				break
			}
		} else {
			continue
		}
	}
	if count == 0 {
		cache.removeOldest() //当没有带过期时间的键值对时，采用LRU策略
	}
}

/*
LRU删除，默认策略，仅删除最后一个
*/

type LRUAll struct{}

func (l *LRUAll) MakeSpace(cache *InMemCache) {
	cache.removeOldest()
}

func NewExpireStrategy(strategy string) ExpireStrategy {
	switch strategy {
	case "RandAll":
		return &RandAll{}
	case "RandVolatile":
		return &RandVolatile{}
	default:
		return &LRUAll{}
	}
}
