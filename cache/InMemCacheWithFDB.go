package cache

import (
	"Godis/config"
	"bufio"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
临时结构体，用于json解析避免[]byte类型解析错误，而临时使用string类型辅助解析
*/
type tempValue struct {
	Val     string
	Created time.Time
	TTL     time.Duration //生存时间
}
type InMemCacheWithFDB struct {
	InMemCache
	fdbDuration int64 //持久化间隔
	fdbDuring   bool  //是否正在做持久化的标志
	//快照用于辅助磁盘存储，键值对位置一一对应
	fileKeys   []string
	fileValues []Value
}

func (cache *InMemCacheWithFDB) FDB() {
	ticker := time.NewTicker(time.Duration(cache.fdbDuration) * time.Second)
	times := ticker.C
	go func() {
		for _ = range times {
			//log.Printf("持久化时间周期 %v\n", t)
			if !cache.fdbDuring {
				cache.copyToFile()
			}
		}
	}()
}

func (cache *InMemCacheWithFDB) copyMem() {
	cache.fileKeys, cache.fileValues = cache.KeysAndValues()
}

//fdb file 格式：
//<keyCount> <keyLen> <valueLen> <key><Value><keyLen> <valueLen> <key><Value>
//例如：
//2 2 3 kyval3 5 keyvalue

//(已解决，原因Get操作需要修改LRU链表，需加写锁) 解决备份文件在没有设置内存限制与数据淘汰时，文件越来越小的问题
func (cache *InMemCacheWithFDB) copyToFile() {
	cache.fdbDuring = true
	defer func() {
		cache.fdbDuring = false
	}()
	//创建备份文件
	file, err := os.OpenFile(config.Config.FDB.FileName+".bak", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("创建备份文件dump.fdb.bak失败\nerror:%v\n", err)
		return
	}

	w := bufio.NewWriter(file)
	start := time.Now()
	//拷贝快照
	cache.copyMem()
	//log.Printf("当前快照keys:%v,当前快照values:%v,当前State:%v", len(cache.fileKeys), len(cache.fileValues), cache.State.Count)
	//写文件
	count := len(cache.fileKeys)
	w.Write([]byte(strconv.Itoa(count) + " "))
	for i := 0; i < count; i++ {
		k := cache.fileKeys[i]
		v := cache.fileValues[i]
		var tempV tempValue
		tempV.Val = string(v.Val)
		tempV.TTL = v.TTL
		tempV.Created = v.Created
		values, _ := json.Marshal(tempV) //value结构体转json存储
		if len(k) == 0 || v.Val == nil {
			continue
		}
		w.Write([]byte(fmt.Sprintf("%v %v %v%v", len(k), len(values), k, string(values))))
	}
	w.Flush()
	cache.fileKeys = cache.fileKeys[0:0] //交由GC回收，清空快照
	cache.fileValues = cache.fileValues[0:0]
	file.Close() //关闭文件描述符，否则无法删除文件并替换
	//备份文件替换
	err = os.Remove(config.Config.FDB.FileName)
	if err != nil {
		log.Printf("remove : 替换备份文件失败,error : %v", err)
		return
	}
	err = os.Rename(config.Config.FDB.FileName+".bak", config.Config.FDB.FileName)
	if err != nil {
		log.Println("rename : 替换备份文件失败")
		return
	}
	log.Printf("备份耗时：%v,备份%v组数据\n", time.Since(start), count)
}

//从FDB文件中加载缓存数据，会同时保存到缓存和快照
//已删除快照，改为当进行fdb时，对数据加读锁，并复制map，释放读锁后，立即对拷贝进行写文件操作。
func (cache *InMemCacheWithFDB) LoadCacheFromFDB() {
	start := time.Now()
	file, err := os.OpenFile(config.Config.FDB.FileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Printf("加载备份文件%s失败\nerror:%v\n", config.Config.FDB.FileName, err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	count, err := readLen(reader)
	if err != nil && err != io.EOF {
		log.Printf("读取%s文件内容失败\nerror:%v\n", config.Config.FDB.FileName, err)
		return
	}
	//cache.lock.Lock()
	//defer cache.lock.Unlock()
	cache.memCache = make(map[string]*list.Element, count)
	for i := 1; i <= count; i++ {
		keyLen, err := readLen(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("%s文件内容格式不正确或已损坏\nerror:%v\n", config.Config.FDB.FileName, err)
			return
		}
		valueLen, err := readLen(reader)
		if err != nil {
			log.Printf("%s文件内容格式不正确或已损坏\nerror:%v\n", config.Config.FDB.FileName, err)
			return
		}
		key := make([]byte, keyLen)
		val := make([]byte, valueLen)
		n, err := io.ReadFull(reader, key)
		if err != nil || n != keyLen {
			log.Printf("读取%s文件内容失败\nerror:%v\n", config.Config.FDB.FileName, err)
			return
		}
		n, err = io.ReadFull(reader, val)
		if err != nil || n != valueLen {
			log.Printf("读取%s文件内容失败\nerror:%v\n", config.Config.FDB.FileName, err)
			return
		}
		var v tempValue
		json.Unmarshal(val, &v)
		cache.Set(string(key), Value{[]byte(v.Val), v.Created, v.TTL})
		//log.Printf("当前count:%v,当前缓存大小%v,",i,len(cache.memCache))
	}
	//（已解决，原因拷贝时数据拷贝不完整，是由于Get操作修改LRU链表需要加写锁） bug:当加载数据很大时，加载的数据量与缓存中的数据量不一致
	//查看dump文件，发现最后很多数据记录为null等默认零值，并且每次发生在get测试后，备份数据就开始出现很多null值
	log.Printf("加载%s文件内容成功，加载了%v组数据，此时缓存中数据%v，加载耗时%v\n", config.Config.FDB.FileName, count, cache.State.Count, time.Since(start))
}

func readLen(reader *bufio.Reader) (int, error) {
	readString, err := reader.ReadString(' ')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(readString))
}

func NewInMemCacheWithFDB(fdbDuration int64, memoryThreshold int64, expireCycle time.Duration, strategy ExpireStrategy) *InMemCacheWithFDB {
	mem := NewInMemCacheWithMemoryThreshold(memoryThreshold, expireCycle, strategy)
	return &InMemCacheWithFDB{
		InMemCache:  *mem,
		fdbDuration: fdbDuration,
		fdbDuring:   false,
		//fileCache 不初始化，当第一次备份时复制memCache
		//fileCache:   make(map[string][]byte),
	}
}
