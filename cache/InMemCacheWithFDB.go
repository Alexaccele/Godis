package cache

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InMemCacheWithFDB struct {
	InMemCache
	fdbDuration int64            //持久化间隔
	fdbDuring 	bool            //是否正在做持久化的标志
	fileCache   map[string]value //快照等同于备份文件内容
}

func (cache *InMemCacheWithFDB) FDB() {
	ticker := time.NewTicker(time.Duration(cache.fdbDuration) * time.Second)
	times := ticker.C
	go func() {
		for t := range times {
			log.Printf("持久化时间周期 %v\n", t)
			if !cache.fdbDuring{
				cache.copyToFile()
			}
		}
	}()
}

func (cache *InMemCacheWithFDB) copyMem(){
	cache.fileCache = make(map[string]value)
	for k,v := range cache.memCache{
		cache.fileCache[k] = v;
	}
}
//fdb file 格式：
//<keyCount> <keyLen> <valueLen> <key><value><keyLen> <valueLen> <key><value>
//例如：
//2 2 3 kyval3 5 keyvalue
//TODO 后期考虑增量替换，提高持久化性能
func (cache *InMemCacheWithFDB) copyToFile() {
	cache.lock.RLock()
	cache.fdbDuring = true
	defer func() {
		cache.fdbDuring = false
	}()
	//创建备份文件
	file, err := os.OpenFile("dump.fdb.bak", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("创建备份文件dump.fdb.bak失败\nerror:%v\n", err)
		return
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	start := time.Now()
	//拷贝快照
	cache.copyMem()
	cache.lock.RUnlock()
	//写文件
	count := len(cache.fileCache)
	w.Write([]byte(strconv.Itoa(count) + " "))
	for k, v := range cache.fileCache {
		values, _ := json.Marshal(v)//value结构体转json存储
		w.Write([]byte(fmt.Sprintf("%v %v %v%v", len(k), len(values),k, string(values))))
	}
	w.Flush()
	cache.fileCache = nil //交由GC回收，清空快照
	//备份文件替换
	err = os.Remove("dump.fdb")
	if err!=nil{
		log.Println("替换备份文件失败")
		return
	}
	err = os.Rename("dump.fdb.bak", "dump.fdb")
	if err!=nil{
		log.Println("替换备份文件失败")
		return
	}
	log.Printf("备份耗时：%v\n", time.Since(start))
}

//从FDB文件中加载缓存数据，会同时保存到缓存和快照
//已删除快照，改为当进行fdb时，对数据加读锁，并复制map，释放读锁后，立即对拷贝进行写文件操作。
func (cache *InMemCacheWithFDB) LoadCacheFromFDB() {
	file, err := os.OpenFile("dump.fdb", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("加载备份文件dump.fdb失败\nerror:%v\n", err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	count, err := readLen(reader)
	if err != nil && err != io.EOF {
		log.Printf("读取dump.fdb文件内容失败\nerror:%v\n", err)
		return
	}
	cache.lock.Lock()
	defer cache.lock.Unlock()
	cache.memCache = make(map[string]value, count)
	for i := count; i >= 0; i-- {
		keyLen, err := readLen(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("dump.fdb文件内容格式不正确或已损坏\nerror:%v\n", err)
			return
		}
		valueLen, err := readLen(reader)
		if err != nil {
			log.Printf("dump.fdb文件内容格式不正确或已损坏\nerror:%v\n", err)
			return
		}
		key := make([]byte, keyLen)
		val := make([]byte, valueLen)
		n, err := io.ReadFull(reader, key)
		if err != nil || n != keyLen {
			log.Printf("读取dump.fdb文件内容失败\nerror:%v\n", err)
			return
		}
		n, err = io.ReadFull(reader, val)
		if err != nil || n != valueLen {
			log.Printf("读取dump.fdb文件内容失败\nerror:%v\n", err)
			return
		}
		var v value
		json.Unmarshal(val,&v)
		cache.memCache[string(key)] = v
	}
	log.Printf("加载dump.fdb文件内容成功，加载了%v组数据\n", count)
}

func readLen(reader *bufio.Reader) (int, error) {
	readString, err := reader.ReadString(' ')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(readString))
}

func NewInMemCacheWithFDB(fdbDuration int64) *InMemCacheWithFDB {
	return &InMemCacheWithFDB{
		InMemCache: InMemCache{
			lock:     sync.RWMutex{},
			memCache: make(map[string]value),
			State:    State{},
		},
		fdbDuration: fdbDuration,
		fdbDuring:	false,
		//fileCache 不初始化，当第一次备份时复制memCache
		//fileCache:   make(map[string][]byte),
	}
}
