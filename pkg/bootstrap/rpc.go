package bootstrap

import (
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"im/pkg/grpcware"
)

// MustRPCServer 创建 gRPC 服务并挂载业务访问日志拦截器。
func MustRPCServer(c zrpc.RpcServerConf, register func(*grpc.Server)) *zrpc.RpcServer {
	s := zrpc.MustNewServer(c, register)
	s.AddUnaryInterceptors(grpcware.UnaryAccessLog)
	return s
}
