package main

import (
	"flag"
	"fmt"

	"im/apps/notification/rpc/internal/config"
	"im/apps/notification/rpc/internal/server"
	"im/apps/notification/rpc/internal/svc"
	"im/pkg/bootstrap"
	"im/apps/notification/rpc/notification"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/notification.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	bootstrap.SilenceZeroNoise()
	ctx := svc.NewServiceContext(c)

	s := bootstrap.MustRPCServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		notification.RegisterNotificationServer(grpcServer, server.NewNotificationServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
