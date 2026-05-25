package svc

import (
	"database/sql"

	"im/apps/cron/internal/config"
	"im/pkg/redisclient"
	"im/pkg/rocketmq"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	DB          *sql.DB
	Redis       *redisclient.Client
	GatewayPush *rocketmq.Producer
	Producer    *rocketmq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	producer := rocketmq.MustProducer(c.RocketMQ.NameServer)
	return &ServiceContext{
		Config:      c,
		DB:          db,
		Redis:       redisclient.New(c.Redis.Addr),
		GatewayPush: producer,
		Producer:    producer,
	}
}

func (s *ServiceContext) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
	if s.Producer != nil {
		_ = s.Producer.Close()
	}
}
