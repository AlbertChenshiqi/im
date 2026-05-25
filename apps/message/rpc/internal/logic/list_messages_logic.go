package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/message/rpc/internal/svc"
	"im/apps/message/rpc/message"

)

type ListMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListMessagesLogic) ListMessages(in *message.ListMessagesReq) (*message.ListMessagesResp, error) {
	msgs, err := l.svcCtx.MessageRepo.ListMessages(l.ctx, in.ConvId, in.BeforeSeq, int(in.Limit))
	if err != nil {
		return nil, err
	}
	out := make([]*message.MessageItem, 0, len(msgs))
	for _, m := range msgs {
		parts := make([]*message.InputPart, len(m.Input))
		for i, it := range m.Input {
			parts[i] = &message.InputPart{MsgType: it.MsgType, Content: it.Content}
		}
		out = append(out, &message.MessageItem{
			Id: m.ID, ConvId: m.ConvID, SenderId: m.SenderID, Seq: m.Seq, Input: parts,
		})
	}
	return &message.ListMessagesResp{Messages: out}, nil
}
