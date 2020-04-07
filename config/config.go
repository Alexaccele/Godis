package config

import (
	"time"
)

//服务类型配置
type ServiceConfig struct {
	ServiceType string `toml:"service_type"`
	TcpPort     string `toml:"tcp-port"`
	HttpPort    string `toml:"http-port"`
}

//过期策略配置
type ExpireStrategyConfig struct {
	Strategy          string
	MemoryThreshold   int64
	ExpireCycle       time.Duration
	DefaultExpireTime int64
}

//持久化配置
type FDBConfig struct {
	FDBDuration int64
	FileName    string
}

//分布式节点配置
type NodeConifg struct {
	Node    string
	Cluster string
}

type GlobalConfig struct {
	Service        ServiceConfig
	ExpireStrategy ExpireStrategyConfig
	FDB            FDBConfig
	Node           NodeConifg
}

var Config GlobalConfig
