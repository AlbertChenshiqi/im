package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/notification/api/internal/svc"
	"im/apps/notification/api/internal/types"
	"im/pkg/jwtx"
)

type MarkReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *MarkReadLogic) MarkRead(req *types.IdPathReq) (*types.StatusResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	_ = l.svcCtx.NotificationRepo.MarkRead(l.ctx, uid, req.Id)
	return &types.StatusResp{Status: "ok"}, nil
}
