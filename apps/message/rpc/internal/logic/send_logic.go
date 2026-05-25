package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/message/rpc/internal/svc"
	"im/apps/message/rpc/message"
	"im/pkg/models"
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
	input := make([]models.MessageInput, len(in.Input))
	for i, p := range in.Input {
		if p == nil {
			continue
		}
		input[i] = models.MessageInput{MsgType: p.MsgType, Content: p.Content}
	}
	res, err := l.svcCtx.Sender.Send(l.ctx, msgcore.SendInput{
		SenderID: in.SenderId, ConvID: in.ConvId, Input: input,
		ClientMsgID: in.ClientMsgId, BizSeq: in.BizSeq, SendTs: in.SendTs, ServerRecvMs: in.ServerRecvMs,
	})
	if err != nil {
		l.Errorf("[message] send failed sender=%d conv=%s err=%v", in.SenderId, in.ConvId, err)
		return nil, err
	}
	l.Infof("[message] send ok sender=%d conv=%s msg_id=%d seq=%d", in.SenderId, in.ConvId, res.MsgID, res.Seq)
	return &message.SendResp{MsgId: res.MsgID, Seq: res.Seq}, nil
}
