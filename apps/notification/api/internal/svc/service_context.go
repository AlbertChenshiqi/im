package svc

import (
	"context"

	"im/apps/notification/api/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config           config.Config
	NotificationRepo *repo.NotificationRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{Config: c, NotificationRepo: repo.NewNotificationRepo(pool)}
}
