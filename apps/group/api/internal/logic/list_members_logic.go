package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/api/internal/svc"
	"im/apps/group/api/internal/types"
)

type ListMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMembersLogic {
	return &ListMembersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListMembersLogic) ListMembers(req *types.ListMembersReq) (*types.ListMembersResp, error) {
	ids, next, err := l.svcCtx.GroupRepo.ListMembers(l.ctx, req.Id, req.Cursor, 500)
	if err != nil {
		return nil, err
	}
	return &types.ListMembersResp{UserIds: ids, NextCursor: next}, nil
}
