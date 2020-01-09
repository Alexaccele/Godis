package main

import (
	"Godis/cache/impl"
	"Godis/http"
	"flag"
	"fmt"
)

var(
	p string
)

func init()  {
	flag.StringVar(&p,"p","9090","服务监听端口")
}
func main()  {
	flag.Parse()
	server := http.NewServer(impl.NewInMemCache())
	server.Listen(fmt.Sprintf(":%v",p))
}
