package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"im/apps/gateway/api/internal/config"
	"im/apps/gateway/api/internal/handler"
	"im/apps/gateway/api/internal/hub"
	"im/apps/gateway/api/internal/push"
	"im/apps/gateway/api/internal/svc"
	"im/pkg/bootstrap"
)

var configFile = flag.String("f", "etc/gateway-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	bootstrap.SilenceZeroNoise()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := rest.MustNewServer(c.RestConf)
	bootstrap.UseRESTAccessLog(server)
	defer server.Stop()

	svcCtx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, svcCtx)

	h := hub.New(svcCtx)
	var pushWG sync.WaitGroup
	consumer := push.NewConsumer(c.Kafka.Brokers, h)
	consumer.Start(ctx, &pushWG)

	server.AddRoute(rest.Route{Method: http.MethodGet, Path: "/v1/ws", Handler: handler.WSHandler(svcCtx, h)})

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
		server.Stop()
	}()

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
	pushWG.Wait()
}
