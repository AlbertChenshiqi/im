package svc

import (

	"im/apps/conversation/rpc/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config   config.Config
	ConvRepo *repo.ConversationRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{Config: c, ConvRepo: repo.NewConversationRepo(db)}
}
