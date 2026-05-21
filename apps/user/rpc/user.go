package main

import (
	"flag"
	"fmt"

	"im/apps/user/rpc/internal/config"
	"im/apps/user/rpc/internal/server"
	"im/apps/user/rpc/internal/svc"
	"im/apps/user/rpc/user"
	"im/pkg/bootstrap"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	bootstrap.SilenceZeroNoise()
	ctx := svc.NewServiceContext(c)

	s := bootstrap.MustRPCServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
