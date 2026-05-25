package svc

import (

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
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{
		Config:           c,
		NotificationRepo: repo.NewNotificationRepo(db),
		Producer:         rocketmq.MustProducer(c.RocketMQ.NameServer),
	}
}
