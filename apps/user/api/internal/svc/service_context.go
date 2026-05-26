package svc

import (
	"log"

	"im/apps/user/api/internal/config"
	"im/pkg/redisclient"
	"im/pkg/repo"
)

type ServiceContext struct {
	Config   config.Config
	UserRepo *repo.UserRepo
	Redis    *redisclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := repo.NewPool(c.MySQL.DSN)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	return &ServiceContext{
		Config:   c,
		UserRepo: repo.NewUserRepo(db),
		Redis:    redisclient.New(c.Redis.Addr),
	}
}
