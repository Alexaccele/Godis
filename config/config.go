package config

import (
	"github.com/BurntSushi/toml"
	"log"
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
}

type GlobalConfig struct {
	Service        ServiceConfig
	ExpireStrategy ExpireStrategyConfig
	FDB            FDBConfig
}

var Config GlobalConfig

func init() {
	if _, err := toml.DecodeFile("config.toml", &Config); err != nil {
		log.Fatalf("读取配置文件config.toml失败,error:%v", err)
	}
}
