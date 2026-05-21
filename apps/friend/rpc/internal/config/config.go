package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Postgres struct{ DSN string }
	ConversationRpc struct {
		Endpoints []string
	}
}
