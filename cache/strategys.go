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
type RandAll struct {}

func (r *RandAll) MakeSpace(cache *InMemCache)  {
	toDelCount := 10
	count := 0
	for k,_ := range cache.memCache{//golang 实现的for-range map是随机化的，故直接遍历
		if count < toDelCount{
			cache.Del(k)
			count++
		}
		break
	}
}

/*
从所有带过期时间的键值对中随机删除
*/
type RandVolatile struct {}

func (r *RandVolatile) MakeSpace(cache *InMemCache)  {
	toDelCount := 10
	count := 0
	for k,v := range cache.memCache{
		if v.TTL > 0{
			if count < toDelCount{
				cache.Del(k)
				count++
			}else{
				break
			}
		}else{
			continue
		}
	}
}
