package logic

import (
	"context"

	"im/apps/gateway/api/internal/hub"
	"im/apps/gateway/api/internal/protocol"
)

type WSPingLogic struct {
	ctx context.Context
	h   *hub.Hub
}

func NewWSPingLogic(ctx context.Context, h *hub.Hub) *WSPingLogic {
	return &WSPingLogic{ctx: ctx, h: h}
}

func (l *WSPingLogic) Ping(_ *hub.Session) protocol.PongOut {
	return protocol.NewPong()
}
