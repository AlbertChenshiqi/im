package svc

import (
	"context"

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
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{
		Config:   c,
		ConvRepo: repo.NewConversationRepo(pool),
		Redis:    redisclient.New(c.Redis.Addr),
	}
}
