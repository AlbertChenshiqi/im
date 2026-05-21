package svc

import (
	"context"
	"log"

	"im/pkg/repo"
	"im/apps/user/api/internal/config"
)

type ServiceContext struct {
	Config   config.Config
	UserRepo *repo.UserRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool, err := repo.NewPool(context.Background(), c.Postgres.DSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	return &ServiceContext{
		Config:   c,
		UserRepo: repo.NewUserRepo(pool),
	}
}
