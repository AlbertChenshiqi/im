package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/zeromicro/go-zero/core/conf"

	"im/apps/transfer/internal/config"
	"im/apps/transfer/internal/runner"
	"im/apps/transfer/internal/svc"
	"im/pkg/bootstrap"
)

var configFile = flag.String("f", "etc/transfer.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	bootstrap.SilenceZeroNoise()

	port := c.HealthPort
	if port == 0 {
		port = 10800
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svcCtx := svc.NewServiceContext(c)
	defer svcCtx.Close()

	var wg sync.WaitGroup
	runner.StartAll(ctx, svcCtx, &wg)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	go func() {
		addr := fmt.Sprintf(":%d", port)
		log.Printf("[transfer] health listening on %s", addr)
		_ = http.ListenAndServe(addr, mux)
	}()

	log.Println("[transfer] all async tasks started")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("[transfer] shutting down...")
	cancel()
	wg.Wait()
}
