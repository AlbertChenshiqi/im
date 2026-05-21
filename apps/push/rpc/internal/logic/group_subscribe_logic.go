package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/push/rpc/internal/svc"
	"im/apps/push/rpc/push"
)

type GroupSubscribeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSubscribeLogic {
	return &GroupSubscribeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupSubscribeLogic) GroupSubscribe(in *push.GroupSubscribeReq) (*push.GroupSubscribeResp, error) {
	_ = l.svcCtx.Redis.SetOnline(l.ctx, in.UserId, 300)
	return &push.GroupSubscribeResp{}, nil
}
