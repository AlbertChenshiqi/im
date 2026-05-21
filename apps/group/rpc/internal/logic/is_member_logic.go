package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/rpc/group"
	"im/apps/group/rpc/internal/svc"
)

type IsMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsMemberLogic {
	return &IsMemberLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *IsMemberLogic) IsMember(in *group.IsMemberReq) (*group.IsMemberResp, error) {
	ok, muted, err := l.svcCtx.GroupRepo.IsMember(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		return nil, err
	}
	return &group.IsMemberResp{Ok: ok, Muted: muted}, nil
}
