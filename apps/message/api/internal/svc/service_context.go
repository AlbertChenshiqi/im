package svc

import (
	"context"

	"im/apps/message/api/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	MessageRepo *repo.MessageRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{
		Config:      c,
		MessageRepo: repo.NewMessageRepo(pool),
	}
}
