package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/gateway/api/internal/hub"
	"im/apps/gateway/api/internal/protocol"
	"im/apps/gateway/api/internal/svc"
	"im/apps/message/rpc/message"
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
	out, err := l.svcCtx.MessageRpc.Send(l.ctx, &message.SendReq{
		SenderId:    session.UID(),
		ConvId:      frame.ConvId,
		Content:     frame.Content,
		MsgType:     frame.MsgType,
		ClientMsgId: frame.ClientMsgId,
	})
	if err != nil {
		logx.WithContext(l.ctx).Errorf("[gateway] ws send failed uid=%d conv=%s err=%v", session.UID(), frame.ConvId, err)
		e := protocol.NewErrorOut(code.GatewaySendFailed, err.Error())
		return protocol.SentOut{}, &e
	}
	logx.WithContext(l.ctx).Infof("[gateway] ws send ok uid=%d conv=%s msg_id=%d seq=%d", session.UID(), frame.ConvId, out.MsgId, out.Seq)
	return protocol.NewSent(out.MsgId, out.Seq), nil
}
