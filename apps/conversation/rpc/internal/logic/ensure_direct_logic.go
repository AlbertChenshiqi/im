package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/conversation/rpc/conversation"
	"im/apps/conversation/rpc/internal/svc"
)

type EnsureDirectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnsureDirectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnsureDirectLogic {
	return &EnsureDirectLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *EnsureDirectLogic) EnsureDirect(in *conversation.EnsureDirectReq) (*conversation.EnsureDirectResp, error) {
	cid, err := l.svcCtx.ConvRepo.EnsureDirect(l.ctx, in.UserA, in.UserB)
	if err != nil {
		return nil, err
	}
	return &conversation.EnsureDirectResp{ConvId: cid}, nil
}
