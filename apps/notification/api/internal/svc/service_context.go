package svc

import (

	"im/apps/notification/api/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config           config.Config
	NotificationRepo *repo.NotificationRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{Config: c, NotificationRepo: repo.NewNotificationRepo(db)}
}
