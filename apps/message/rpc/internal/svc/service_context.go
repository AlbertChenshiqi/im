package svc

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"

	"im/apps/group/rpc/group_client"
	"im/apps/message/rpc/internal/config"
	"im/pkg/kafka"
	"im/pkg/msgcore"
	"im/pkg/redisclient"
	"im/pkg/repo"
	"im/pkg/snowflake"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config      config.Config
	MessageRepo *repo.MessageRepo
	Sender      *msgcore.Sender
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	rdb := redisclient.New(c.RedisStore.Addr)
	w := kafka.NewWriter(c.Kafka.Brokers, "im.message.send")
	var groupRpc group_client.Group
	if len(c.GroupRpc.Endpoints) > 0 {
		groupRpc = group_client.NewGroup(zrpc.MustNewClient(zrpc.RpcClientConf{Endpoints: c.GroupRpc.Endpoints, NonBlock: true}))
	}
	convRepo := repo.NewConversationRepo(pool)
	return &ServiceContext{
		Config:      c,
		MessageRepo: repo.NewMessageRepo(pool),
		Sender: &msgcore.Sender{
			RDB: rdb, Writer: w, SF: snowflake.New(5), GroupRpc: groupRpc, ConvRepo: convRepo,
		},
	}
}
