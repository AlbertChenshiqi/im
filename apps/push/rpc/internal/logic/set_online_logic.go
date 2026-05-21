package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/push/rpc/internal/svc"
	"im/apps/push/rpc/push"
)

type SetOnlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetOnlineLogic {
	return &SetOnlineLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SetOnlineLogic) SetOnline(in *push.SetOnlineReq) (*push.SetOnlineResp, error) {
	_ = l.svcCtx.Redis.SetOnline(l.ctx, in.UserId, 300)
	return &push.SetOnlineResp{}, nil
}
