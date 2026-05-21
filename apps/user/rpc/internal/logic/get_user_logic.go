package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/rpc/internal/convert"
	"im/apps/user/rpc/internal/svc"
	"im/apps/user/rpc/user"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetUserLogic) GetUser(in *user.GetUserReq) (*user.GetUserResp, error) {
	u, err := l.svcCtx.UserRepo.GetByID(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &user.GetUserResp{User: convert.UserToRPC(u)}, nil
}
