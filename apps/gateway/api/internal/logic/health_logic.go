package logic

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/gateway/api/internal/svc"
	"im/apps/gateway/api/internal/types"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *HealthLogic) Health() (*types.HealthResp, error) {
	if err := l.svcCtx.Redis.Ping(l.ctx); err != nil {
		return &types.HealthResp{Status: "degraded"}, fmt.Errorf("redis: %w", err)
	}
	return &types.HealthResp{Status: "ok"}, nil
}
