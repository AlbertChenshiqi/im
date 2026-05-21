package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/message/rpc/internal/svc"
	"im/apps/message/rpc/message"
	"im/pkg/msgcore"
)

type SendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendLogic {
	return &SendLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SendLogic) Send(in *message.SendReq) (*message.SendResp, error) {
	res, err := l.svcCtx.Sender.Send(l.ctx, msgcore.SendInput{
		SenderID: in.SenderId, ConvID: in.ConvId, Content: in.Content,
		MsgType: in.MsgType, ClientMsgID: in.ClientMsgId,
	})
	if err != nil {
		l.Errorf("[message] send failed sender=%d conv=%s err=%v", in.SenderId, in.ConvId, err)
		return nil, err
	}
	l.Infof("[message] send ok sender=%d conv=%s msg_id=%d seq=%d", in.SenderId, in.ConvId, res.MsgID, res.Seq)
	return &message.SendResp{MsgId: res.MsgID, Seq: res.Seq}, nil
}
