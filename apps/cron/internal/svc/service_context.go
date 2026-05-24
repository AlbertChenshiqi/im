package svc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/apps/cron/internal/config"
	"im/pkg/redisclient"
	"im/pkg/rocketmq"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	Pool        *pgxpool.Pool
	Redis       *redisclient.Client
	GatewayPush *rocketmq.Producer
	Producer    *rocketmq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	producer := rocketmq.MustProducer(c.RocketMQ.NameServer)
	return &ServiceContext{
		Config:      c,
		Pool:        pool,
		Redis:       redisclient.New(c.Redis.Addr),
		GatewayPush: producer,
		Producer:    producer,
	}
}

func (s *ServiceContext) Close() {
	if s.Pool != nil {
		s.Pool.Close()
	}
	if s.Producer != nil {
		_ = s.Producer.Close()
	}
}
