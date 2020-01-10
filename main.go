package main

import (
	"Godis/cache/impl"
	"Godis/http"
	"Godis/tcp"
	"flag"
	"fmt"
)

var(
	p string
	s string
)

func init()  {
	flag.StringVar(&p,"p","9090","服务监听端口")
	flag.StringVar(&s,"s","tcp","服务协议方式")
}
func main()  {
	flag.Parse()
	switch s {
	case "tcp":
		tcp.NewServer(impl.NewInMemCache()).Listen(p)//tcp服务,默认的服务方式，比HTTP效率高
	case "http":
		http.NewServer(impl.NewInMemCache()).Listen(p)//http服务
	default:
		fmt.Errorf("未支持服务类型 %v\n",s)
		return
	}
}
