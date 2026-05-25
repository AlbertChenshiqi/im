package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	// JwtAuth 避免与 zrpc.RpcServerConf.Auth（bool 开关）重名
	JwtAuth struct {
		AccessSecret string
	}
	MySQL struct {
		DSN string
	}
}
