package cluster

import (
	"Godis/consistent"
	"github.com/hashicorp/memberlist"
	"io/ioutil"
	"time"
)

type Node interface {
	Members() []string
	Addr() string
	ShouldProcess(key string) (string, bool)
}

type node struct {
	*consistent.Consistent
	addr string //包含地址与端口
}

func (n *node) Addr() string {
	return n.addr
}

func (n *node) ShouldProcess(key string) (string, bool) {
	addr, _ := n.Get(key)
	//log.Printf("addr:%v,n.addr:%v\n",addr,n.addr)
	return addr, addr == n.addr
}

//基于gossip协议，创建一致性哈希环
func New(addr, cluster, httpPort string) (Node, error) {
	config := memberlist.DefaultLANConfig()
	config.Name = addr + ":" + httpPort //将http的端口信息包含进去，方便后续取出
	config.BindAddr = addr
	config.LogOutput = ioutil.Discard //抛弃日志
	c, err := memberlist.Create(config)
	if err != nil {
		return nil, err
	}
	if cluster == "" {
		cluster = addr
	}
	clu := []string{cluster}
	_, err = c.Join(clu)
	if err != nil {
		return nil, err
	}
	circle := consistent.New()
	circle.NumberOfReplicas = 256
	go func() {
		for {
			members := c.Members()
			nodes := make([]string, len(members))
			for i, n := range members {
				nodes[i] = n.Name
			}
			circle.Set(nodes)
			time.Sleep(time.Second)
		}
	}()
	return &node{circle, addr + ":" + httpPort}, nil
}
