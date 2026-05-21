package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
	"im/pkg/jwtx"
)

type UpsertDeviceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpsertDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertDeviceLogic {
	return &UpsertDeviceLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpsertDeviceLogic) UpsertDevice(req *types.DeviceReq) (*types.StatusResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	if uid == 0 {
		return nil, code.New(code.CommonUnauthorized)
	}
	if req.DeviceId == "" {
		return nil, code.New(code.UserDeviceRequired)
	}
	if err := l.svcCtx.UserRepo.UpsertDevice(l.ctx, uid, req.DeviceId, req.PushToken, req.Platform); err != nil {
		return nil, err
	}
	return &types.StatusResp{Status: "ok"}, nil
}
