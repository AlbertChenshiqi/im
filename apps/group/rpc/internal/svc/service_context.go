package svc

import (

	"im/apps/group/rpc/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config    config.Config
	GroupRepo *repo.GroupRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	return &ServiceContext{Config: c, GroupRepo: repo.NewGroupRepo(db)}
}
