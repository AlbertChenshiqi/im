package svc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"

	"im/apps/cron/internal/config"
	"im/pkg/events"
	imkafka "im/pkg/kafka"
	"im/pkg/redisclient"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	Pool        *pgxpool.Pool
	Redis       *redisclient.Client
	GatewayPush *kafkago.Writer
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{
		Config:      c,
		Pool:        pool,
		Redis:       redisclient.New(c.Redis.Addr),
		GatewayPush: imkafka.NewWriter(c.Kafka.Brokers, events.TopicGatewayPush),
	}
}

func (s *ServiceContext) Close() {
	if s.Pool != nil {
		s.Pool.Close()
	}
	if s.GatewayPush != nil {
		_ = s.GatewayPush.Close()
	}
}
