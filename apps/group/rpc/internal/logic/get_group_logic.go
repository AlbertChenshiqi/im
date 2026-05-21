package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/rpc/group"
	"im/apps/group/rpc/internal/svc"
)

type GetGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupLogic {
	return &GetGroupLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetGroupLogic) GetGroup(in *group.GetGroupReq) (*group.GetGroupResp, error) {
	g, err := l.svcCtx.GroupRepo.GetGroup(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &group.GetGroupResp{Group: &group.GroupInfo{
		Id: g.ID, Name: g.Name, OwnerId: g.OwnerID, MaxMembers: int32(g.MaxMembers),
	}}, nil
}
