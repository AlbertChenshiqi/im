package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
	"im/pkg/jwtx"
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
	if uid == 0 {
		return nil, code.New(code.CommonUnauthorized)
	}
	ttl := l.svcCtx.Config.OnlineTTLSeconds
	if ttl <= 0 {
		ttl = 300
	}
	_ = l.svcCtx.Redis.SetOnline(l.ctx, uid, ttl)
	return &types.StatusResp{Status: "online"}, nil
}
