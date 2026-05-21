package svc

import (
	"context"

	kafkago "github.com/segmentio/kafka-go"

	"im/apps/notification/rpc/internal/config"
	"im/pkg/kafka"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config           config.Config
	NotificationRepo *repo.NotificationRepo
	NotifyWriter     *kafkago.Writer
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{
		Config:           c,
		NotificationRepo: repo.NewNotificationRepo(pool),
		NotifyWriter:     kafka.NewWriter(c.Kafka.Brokers, "im.notification.system"),
	}
}
