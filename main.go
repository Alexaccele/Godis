package main

import (
	"Godis/cache/impl"
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
	switch s {
	case "tcp":
		cache := impl.NewInMemCacheWithFDB(5)
		cache.LoadCacheFromFDB()
		cache.FDB()
		tcp.NewServer(cache).Listen(tcpPort) //tcp服务,默认的服务方式，比HTTP效率高
	case "http":
		http.NewServer(impl.NewInMemCache()).Listen(httpPort) //http服务
	default:
		fmt.Errorf("未支持服务类型 %v\n", s)
		return
	}
}
