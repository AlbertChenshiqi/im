package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/rpc/group"
	"im/apps/group/rpc/internal/svc"
)

type ListMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMembersLogic {
	return &ListMembersLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListMembersLogic) ListMembers(in *group.ListMembersReq) (*group.ListMembersResp, error) {
	limit := int(in.Limit)
	if limit <= 0 {
		limit = 500
	}
	ids, next, err := l.svcCtx.GroupRepo.ListMembers(l.ctx, in.GroupId, in.Cursor, limit)
	if err != nil {
		return nil, err
	}
	return &group.ListMembersResp{UserIds: ids, NextCursor: next}, nil
}
