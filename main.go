package main

import (
	"Godis/cache"
	"Godis/config"
	"Godis/http"
	"Godis/tcp"
	"flag"
	"fmt"
)

var (
	httpPort string
	tcpPort  string
	s        string
)

func init() {
	flag.StringVar(&httpPort, "http-port", "9090", "HTTP服务监听端口")
	flag.StringVar(&tcpPort, "tcp-port", "2333", "TCP服务监听端口")
	flag.StringVar(&s, "s", "tcp", "服务协议方式")
}
func main() {
	flag.Parse()
	cache := cache.NewInMemCacheWithFDB(config.Config.FDB.FDBDuration,
		1<<20*config.Config.ExpireStrategy.MemoryThreshold,
		config.Config.ExpireStrategy.ExpireCycle,
		cache.NewExpireStrategy(config.Config.ExpireStrategy.Strategy)) //持久化周期5s,0表示无内存限制，30s默认检查过期时间
	cache.LoadCacheFromFDB()
	cache.FDB()
	if cache.ExpireCycle > 0 {
		go cache.Expirer()
	}
	switch config.Config.Service.ServiceType {
	case "tcp":
		tcp.NewServer(cache).Listen(config.Config.Service.Port) //tcp服务,默认的服务方式，比HTTP效率高
	case "http":
		http.NewServer(cache).Listen(config.Config.Service.Port) //http服务
	default:
		fmt.Errorf("未支持服务类型 %v\n", s)
		return
	}
}
