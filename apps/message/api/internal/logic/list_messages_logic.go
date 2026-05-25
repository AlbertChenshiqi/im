package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/message/api/internal/svc"
	"im/apps/message/api/internal/types"

)

type ListMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMessagesLogic) ListMessages(req *types.ListMessagesReq) (*types.ListMessagesResp, error) {
	msgs, err := l.svcCtx.MessageRepo.ListMessages(l.ctx, req.Id, req.BeforeSeq, req.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]types.Message, 0, len(msgs))
	for _, m := range msgs {
		parts := make([]types.MessageInput, len(m.Input))
		for i, it := range m.Input {
			parts[i] = types.MessageInput{MsgType: it.MsgType, Content: it.Content}
		}
		out = append(out, types.Message{
			Id: m.ID, ConvId: m.ConvID, SenderId: m.SenderID, Seq: m.Seq, Input: parts,
		})
	}
	return &types.ListMessagesResp{Messages: out}, nil
}
