package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"im/apps/push/rpc/internal/config"
	"im/apps/push/rpc/internal/server"
	"im/apps/push/rpc/internal/svc"
	"im/pkg/bootstrap"
	"im/apps/push/rpc/push"
)

var configFile = flag.String("f", "etc/push.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	bootstrap.SilenceZeroNoise()
	ctx := svc.NewServiceContext(c)

	s := bootstrap.MustRPCServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		push.RegisterPushServer(grpcServer, server.NewPushServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting push rpc at %s...\n", c.ListenOn)
	s.Start()
}
