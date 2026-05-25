package svc

import (

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
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{
		Config:    c,
		GroupRepo: repo.NewGroupRepo(db),
		UserRepo:  repo.NewUserRepo(db),
	}
}
