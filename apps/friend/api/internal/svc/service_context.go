package svc

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"

	"im/apps/conversation/rpc/conversation_client"
	"im/apps/friend/api/internal/config"
	"im/pkg/repo"
	"im/pkg/zerokit"
)

type ServiceContext struct {
	Config           config.Config
	FriendRepo       *repo.FriendRepo
	ConversationRpc  conversation_client.Conversation
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := zerokit.MustPGPool(context.Background(), c.Postgres.DSN)
	var conv conversation_client.Conversation
	if len(c.ConversationRpc.Endpoints) > 0 {
		conv = conversation_client.NewConversation(zrpc.MustNewClient(zrpc.RpcClientConf{Endpoints: c.ConversationRpc.Endpoints, NonBlock: true}))
	}
	return &ServiceContext{
		Config:          c,
		FriendRepo:      repo.NewFriendRepo(pool),
		ConversationRpc: conv,
	}
}
