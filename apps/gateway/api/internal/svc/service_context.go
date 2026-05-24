package svc

import (
	"github.com/zeromicro/go-zero/zrpc"

	"im/apps/gateway/api/internal/config"
	"im/apps/gateway/api/internal/order"
	"im/apps/message/rpc/message_client"
	"im/pkg/redisclient"
)

type ServiceContext struct {
	Config     config.Config
	MessageRpc message_client.Message
	Redis      *redisclient.Client
	Order      *order.Coordinator
}

func NewServiceContext(c config.Config) *ServiceContext {
	windowMs := c.SendOrder.WindowMs
	if windowMs <= 0 {
		windowMs = 200
	}
	s := &ServiceContext{
		Config: c,
		MessageRpc: message_client.NewMessage(zrpc.MustNewClient(zrpc.RpcClientConf{
			Endpoints: c.MessageRpc.Endpoints, NonBlock: true,
		})),
		Redis: redisclient.New(c.Redis.Addr),
	}
	s.Order = order.NewCoordinator(s.MessageRpc, s.Redis, windowMs)
	return s
}
