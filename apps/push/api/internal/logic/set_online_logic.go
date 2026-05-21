package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/pkg/jwtx"
	"im/apps/push/api/internal/svc"
	"im/apps/push/api/internal/types"
)

type SetOnlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetOnlineLogic {
	return &SetOnlineLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *SetOnlineLogic) SetOnline() (*types.StatusResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	_ = l.svcCtx.Redis.SetOnline(l.ctx, uid, 300)
	return &types.StatusResp{Status: "online"}, nil
}
