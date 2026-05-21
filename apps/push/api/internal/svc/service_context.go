package svc

import (
	"im/pkg/redisclient"
	"im/apps/push/api/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redisclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{Config: c, Redis: redisclient.New(c.Redis.Addr)}
}
