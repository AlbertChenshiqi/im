package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Postgres struct{ DSN string }
	GroupRpc struct {
		Endpoints []string
	}
	Kafka struct {
		Brokers []string
	}
	// RedisStore 避免与 zrpc.RpcServerConf.Redis（go-zero RedisKeyConf，需 Host）重名
	RedisStore struct {
		Addr string
	}
}
