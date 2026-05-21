package svcconfig

import "github.com/zeromicro/go-zero/zrpc"

type PostgresConf struct {
	DSN string
}

type AuthConf struct {
	AccessSecret string
	AccessExpire int64 `json:",optional"`
}

type KafkaConf struct {
	Brokers []string
}

type RedisConf struct {
	Addr string
}

type RpcEndpoints struct {
	Endpoints []string `json:",optional"`
}

func MustClient(c RpcEndpoints) zrpc.Client {
	return zrpc.MustNewClient(zrpc.RpcClientConf{Endpoints: c.Endpoints, NonBlock: true})
}
