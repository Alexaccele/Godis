package main

import (
	"Godis/cache"
	"Godis/cluster"
	"Godis/config"
	"Godis/http"
	"Godis/tcp"
	"context"
	"flag"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	conf       string
	memoryUnit map[string]int64
)

func init() {
	flag.StringVar(&conf, "conf", "config.toml", "指定配置文件")
	memoryUnit = map[string]int64{
		"B":  1 << 0,
		"KB": 1 << 10,
		"MB": 1 << 20,
		"GB": 1 << 30,
	}
}
func main() {
	flag.Parse()
	if _, err := toml.DecodeFile(conf, &config.Config); err != nil {
		log.Fatalf("读取配置文件config.toml失败,error:%v", err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	cache := cache.NewInMemCacheWithFDB(config.Config.FDB.FDBDuration,
		memoryUnit[config.Config.ExpireStrategy.MemoryUnit]*config.Config.ExpireStrategy.MemoryThreshold, //单位 * 阈值
		config.Config.ExpireStrategy.ExpireCycle,
		cache.NewExpireStrategy(config.Config.ExpireStrategy.Strategy)) //持久化周期5s,0表示无内存限制，30s默认检查过期时间
	cache.LoadCacheFromFDB()
	cache.FDB()
	if cache.ExpireCycle > 0 {
		go cache.Expirer()
	}
	node, err := cluster.New(config.Config.Node.Node, config.Config.Node.Cluster, config.Config.Service.HttpPort)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	StartTCPServer(ctx, &wg, cache, node)
	StartHTTPServer(ctx, &wg, cache, node)

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
		cancelFunc()
		wg.Wait()
		os.Exit(0)
	}
}

func StartTCPServer(ctx context.Context, wg *sync.WaitGroup, cache *cache.InMemCacheWithFDB, node cluster.Node) {
	defer wg.Done()
	wg.Add(1)
	go tcp.NewServer(cache, node).Listen(config.Config.Service.TcpPort, ctx)
}

func StartHTTPServer(ctx context.Context, wg *sync.WaitGroup, cache *cache.InMemCacheWithFDB, node cluster.Node) {
	defer wg.Done()
	wg.Add(1)
	go http.NewServer(cache, node).Listen(config.Config.Service.HttpPort, ctx)
}
