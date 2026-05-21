package svc

import (
	"context"

	"im/apps/group/api/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config    config.Config
	GroupRepo *repo.GroupRepo
	UserRepo  *repo.UserRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	return &ServiceContext{
		Config:    c,
		GroupRepo: repo.NewGroupRepo(pool),
		UserRepo:  repo.NewUserRepo(pool),
	}
}
