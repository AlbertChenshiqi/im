package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/api/internal/svc"
	"im/apps/group/api/internal/types"
)

type AddMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMembersLogic {
	return &AddMembersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AddMembersLogic) AddMembers(req *types.AddMembersPathReq) (*types.StatusResp, error) {
	if err := l.svcCtx.GroupRepo.AddMembers(l.ctx, req.Id, req.UserIds); err != nil {
		return nil, err
	}
	return &types.StatusResp{Status: "ok"}, nil
}
