package svc

import (
	"context"

	"im/apps/notification/rpc/internal/config"
	"im/pkg/repo"
	"im/pkg/rocketmq"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config           config.Config
	NotificationRepo *repo.NotificationRepo
	Producer         *rocketmq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{
		Config:           c,
		NotificationRepo: repo.NewNotificationRepo(pool),
		Producer:         rocketmq.MustProducer(c.RocketMQ.NameServer),
	}
}
