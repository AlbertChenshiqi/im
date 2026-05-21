package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/api/internal/svc"
	"im/apps/group/api/internal/types"
)

type GetGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupLogic {
	return &GetGroupLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetGroupLogic) GetGroup(req *types.IdPathReq) (*types.Group, error) {
	g, err := l.svcCtx.GroupRepo.GetGroup(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.Group{Id: g.ID, Name: g.Name, OwnerId: g.OwnerID, MaxMembers: g.MaxMembers, Notice: g.Notice}, nil
}
