package svc

import (

	"github.com/zeromicro/go-zero/zrpc"

	"im/apps/group/rpc/group_client"
	"im/apps/message/rpc/internal/config"
	"im/pkg/msgcore"
	"im/pkg/redisclient"
	"im/pkg/repo"
	"im/pkg/rocketmq"
	"im/pkg/snowflake"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	MessageRepo *repo.MessageRepo
	Sender      *msgcore.Sender
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := zerokit.MustMySQL(c.MySQL.DSN)
	rdb := redisclient.New(c.RedisStore.Addr)
	producer := rocketmq.MustProducer(c.RocketMQ.NameServer)
	var groupRpc group_client.Group
	if len(c.GroupRpc.Endpoints) > 0 {
		groupRpc = group_client.NewGroup(zrpc.MustNewClient(zrpc.RpcClientConf{Endpoints: c.GroupRpc.Endpoints, NonBlock: true}))
	}
	convRepo := repo.NewConversationRepo(db)
	return &ServiceContext{
		Config:      c,
		MessageRepo: repo.NewMessageRepo(db),
		Sender: &msgcore.Sender{
			RDB: rdb, Producer: producer, SF: snowflake.New(5), GroupRpc: groupRpc, ConvRepo: convRepo,
		},
	}
}
