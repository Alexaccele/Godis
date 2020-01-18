package impl

import (
	"Godis/cache"
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type inMemCacheWithFDB struct {
	inMemCache
	fdbDuration uint64            //持久化间隔
	fileCache   map[string][]byte //快照等同于备份文件内容
}

func (cache *inMemCacheWithFDB) FDB() {
	ticker := time.NewTicker(time.Duration(cache.fdbDuration) * time.Second)
	times := ticker.C
	go func() {
		for t := range times {
			fmt.Printf("持久化时间周期 %v\n", t)
			cache.copyToFile()
		}
	}()
}

//fdb file 格式：
//<keyCount> <keyLen> <valueLen> <key><value><keyLen> <valueLen> <key><value>
//例如：
//2 2 3 kyval3 5 keyvalue
//TODO 后期优化点，数据压缩，减小fdb文件大小
//TODO 后期考虑增量替换，提高持久化性能
func (cache *inMemCacheWithFDB) copyToFile() {
	cache.lock.RLock()
	//如果缓存没变过，则不需要备份
	//TODO 判断State是否更为快速，但容忍极低概率的State相同，实际内容存在不同
	if reflect.DeepEqual(cache.fileCache, cache.memCache) { //当数据量巨大时，判断太久
		cache.lock.RUnlock()
		return
	}
	file, err := os.OpenFile("dump.fdb", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("创建或打开备份文件dump.fdb失败\nerror:%v\n", err)
		return
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	start := time.Now()
	for k, v := range cache.memCache {
		key := k
		value := v
		cache.fileCache[key] = value
	}
	cache.lock.RUnlock()
	count := len(cache.fileCache)
	w.Write([]byte(strconv.Itoa(count) + " "))
	for k, v := range cache.fileCache {
		w.Write([]byte(fmt.Sprintf("%v %v %v%v", len(k), len(v), k, string(v))))
	}
	w.Flush()
	fmt.Printf("备份耗时：%v\n", time.Since(start))
}

//从FDB文件中加载缓存数据，会同时保存到缓存和快照
//TODO 后期考虑是否删除快照
func (cache *inMemCacheWithFDB) LoadCacheFromFDB() {
	file, err := os.OpenFile("dump.fdb", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("加载备份文件dump.fdb失败\nerror:%v\n", err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	count, err := readLen(reader)
	if err != nil && err != io.EOF {
		fmt.Printf("读取dump.fdb文件内容失败\nerror:%v\n", err)
		return
	}
	cache.lock.Lock()
	defer cache.lock.Unlock()
	cache.memCache = make(map[string][]byte, count)
	cache.fileCache = make(map[string][]byte, count)
	for i := count; i >= 0; i-- {
		keyLen, err := readLen(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("dump.fdb文件内容格式不正确或已损坏\nerror:%v\n", err)
			return
		}
		valueLen, err := readLen(reader)
		if err != nil {
			fmt.Printf("dump.fdb文件内容格式不正确或已损坏\nerror:%v\n", err)
			return
		}
		key := make([]byte, keyLen)
		value := make([]byte, valueLen)
		n, err := io.ReadFull(reader, key)
		if err != nil || n != keyLen {
			fmt.Printf("读取dump.fdb文件内容失败\nerror:%v\n", err)
			return
		}
		n, err = io.ReadFull(reader, value)
		if err != nil || n != valueLen {
			fmt.Printf("读取dump.fdb文件内容失败\nerror:%v\n", err)
			return
		}
		cache.fileCache[string(key)] = value
		cache.memCache[string(key)] = value
	}
	fmt.Printf("加载dump.fdb文件内容成功，加载了%v组数据\n", count)
}

func readLen(reader *bufio.Reader) (int, error) {
	readString, err := reader.ReadString(' ')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(readString))
}

func NewInMemCacheWithFDB(fdbDuration uint64) *inMemCacheWithFDB {
	return &inMemCacheWithFDB{
		inMemCache: inMemCache{
			lock:     sync.RWMutex{},
			memCache: make(map[string][]byte),
			State:    cache.State{},
		},
		fdbDuration: fdbDuration,
		fileCache:   make(map[string][]byte),
	}
}
