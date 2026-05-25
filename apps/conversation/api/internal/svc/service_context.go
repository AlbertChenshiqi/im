package svc

import (

	"im/apps/conversation/api/internal/config"
	"im/pkg/redisclient"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config   config.Config
	ConvRepo *repo.ConversationRepo
	Redis    *redisclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{
		Config:   c,
		ConvRepo: repo.NewConversationRepo(db),
		Redis:    redisclient.New(c.Redis.Addr),
	}
}
