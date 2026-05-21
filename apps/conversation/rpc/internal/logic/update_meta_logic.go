package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/conversation/rpc/conversation"
	"im/apps/conversation/rpc/internal/svc"
)

type UpdateMetaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMetaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMetaLogic {
	return &UpdateMetaLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateMetaLogic) UpdateMeta(in *conversation.UpdateMetaReq) (*conversation.UpdateMetaResp, error) {
	err := l.svcCtx.ConvRepo.UpdateMeta(l.ctx, in.ConvId, in.Seq, in.MsgId, in.Preview)
	return &conversation.UpdateMetaResp{}, err
}
