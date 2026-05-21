package svc

import (
	"github.com/zeromicro/go-zero/zrpc"

	"im/apps/gateway/api/internal/config"
	"im/apps/message/rpc/message_client"
	"im/pkg/redisclient"
)

type ServiceContext struct {
	Config     config.Config
	MessageRpc message_client.Message
	Redis      *redisclient.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		MessageRpc: message_client.NewMessage(zrpc.MustNewClient(zrpc.RpcClientConf{
			Endpoints: c.MessageRpc.Endpoints, NonBlock: true,
		})),
		Redis: redisclient.New(c.Redis.Addr),
	}
}
