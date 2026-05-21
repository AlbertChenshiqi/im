package svc

import (
	"context"

	"im/apps/conversation/rpc/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config   config.Config
	ConvRepo *repo.ConversationRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{Config: c, ConvRepo: repo.NewConversationRepo(pool)}
}
