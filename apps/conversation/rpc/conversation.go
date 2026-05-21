package main

import (
	"flag"
	"fmt"

	"im/apps/conversation/rpc/conversation"
	"im/apps/conversation/rpc/internal/config"
	"im/apps/conversation/rpc/internal/server"
	"im/apps/conversation/rpc/internal/svc"
	"im/pkg/bootstrap"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/conversation.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	bootstrap.SilenceZeroNoise()
	ctx := svc.NewServiceContext(c)

	s := bootstrap.MustRPCServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		conversation.RegisterConversationServer(grpcServer, server.NewConversationServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
