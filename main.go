package main

import (
	"Godis/cache"
	"Godis/cluster"
	"Godis/config"
	"Godis/http"
	"Godis/tcp"
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	httpPort string
	tcpPort  string
	s        string
	n        string
	clust    string
)

func init() {
	flag.StringVar(&httpPort, config.Config.Service.HttpPort, "9090", "HTTP服务监听端口")
	flag.StringVar(&tcpPort, config.Config.Service.TcpPort, "2333", "TCP服务监听端口")
	flag.StringVar(&s, "s", "tcp", "服务协议方式")
	flag.StringVar(&n, "node", config.Config.Node.Node, "本地服务节点地址")
	flag.StringVar(&clust, "cluster", config.Config.Node.Cluster, "加入的集群节点地址")
}
func main() {
	flag.Parse()
	ctx, cancelFunc := context.WithCancel(context.Background())
	cache := cache.NewInMemCacheWithFDB(config.Config.FDB.FDBDuration,
		(1<<20)*config.Config.ExpireStrategy.MemoryThreshold,
		config.Config.ExpireStrategy.ExpireCycle,
		cache.NewExpireStrategy(config.Config.ExpireStrategy.Strategy)) //持久化周期5s,0表示无内存限制，30s默认检查过期时间
	cache.LoadCacheFromFDB()
	cache.FDB()
	if cache.ExpireCycle > 0 {
		go cache.Expirer()
	}
	node, err := cluster.New(n, clust)
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
	//switch config.Config.Service.ServiceType {
	//case "tcp":
	//	tcp.NewServer(cache).Listen(config.Config.Service.TcpPort) //tcp服务,默认的服务方式，比HTTP效率高
	//case "http":
	//	http.NewServer(cache).Listen(config.Config.Service.HttpPort) //http服务
	//default:
	//	fmt.Errorf("未支持服务类型 %v\n", s)
	//	return
	//}
}

func StartTCPServer(ctx context.Context, wg *sync.WaitGroup, cache *cache.InMemCacheWithFDB, node cluster.Node) {
	defer wg.Done()
	wg.Add(1)
	go tcp.NewServer(cache, node).Listen(tcpPort, ctx)
}

func StartHTTPServer(ctx context.Context, wg *sync.WaitGroup, cache *cache.InMemCacheWithFDB, node cluster.Node) {
	defer wg.Done()
	wg.Add(1)
	go http.NewServer(cache, node).Listen(httpPort, ctx)
}
