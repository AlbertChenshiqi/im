package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/api/internal/convert"
	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetUserLogic) GetUser(req *types.IdPathReq) (*types.User, error) {
	u, err := l.svcCtx.UserRepo.GetByID(l.ctx, req.Id)
	if err != nil {
		return nil, code.New(code.UserNotFound)
	}
	resp := convert.UserToAPI(u)
	return &resp, nil
}
