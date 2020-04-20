package cacheClient

//客户端参数增加过期时间
type Cmd struct {
	Name       string
	Key        string
	Value      string
	ExpireTime string
	Error      error
}

type Client interface {
	Run(*Cmd)
	PipelinedRun([]*Cmd)
}

func New(typ, server, port string) Client {
	if typ == "redis" {
		return newRedisClient(server)
	}
	if typ == "http" {
		return newHTTPClient(server, port)
	}
	if typ == "tcp" {
		return newTCPClient(server, port)
	}
	panic("unknown client type " + typ)
}
