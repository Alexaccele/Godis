package main

import (
	"Godis/cache-benchmark/cacheClient"
	"flag"
	"fmt"
)

var(
	s string
	p string
	op string
	k string
	v string
)

func init() {
	flag.StringVar(&s,"s","localhost","服务器地址")
	flag.StringVar(&p,"p","2333","端口号")
	flag.StringVar(&op,"op","get","操作选项，可以是get/set/del")
	flag.StringVar(&k,"k","","key")
	flag.StringVar(&v,"v","","value")
	flag.Parse()
}

func main()  {
	c := cacheClient.New("tcp", s,p)
	cmd := &cacheClient.Cmd{op,k,v,nil}
	c.Run(cmd)
	if cmd.Error != nil {
		fmt.Println("error:", cmd.Error)
	} else {
		fmt.Println(cmd.Value)
	}
}
