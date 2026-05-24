package logic

import (
	"context"

	"im/apps/gateway/api/internal/hub"
	"im/apps/gateway/api/internal/protocol"
	"im/apps/gateway/api/internal/svc"
	"im/pkg/code"
)

type WSSendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWSSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WSSendLogic {
	return &WSSendLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *WSSendLogic) Send(frame protocol.InFrame, session *hub.Session) (protocol.SentOut, *protocol.ErrorOut) {
	if !session.IsAuthed() {
		e := protocol.NewErrorOut(code.GatewayNotAuthed)
		return protocol.SentOut{}, &e
	}
	return l.svcCtx.Order.Submit(l.ctx, frame, session.UID())
}
