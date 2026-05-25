package svc

import (

	"im/apps/message/api/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	MessageRepo *repo.MessageRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{
		Config:      c,
		MessageRepo: repo.NewMessageRepo(db),
	}
}
