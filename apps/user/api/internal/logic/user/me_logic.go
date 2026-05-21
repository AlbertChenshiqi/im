package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/api/internal/convert"
	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
	"im/pkg/jwtx"
)

type MeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MeLogic {
	return &MeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *MeLogic) Me() (*types.User, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	if uid == 0 {
		return nil, code.New(code.CommonUnauthorized)
	}
	u, err := l.svcCtx.UserRepo.GetByID(l.ctx, uid)
	if err != nil {
		return nil, code.New(code.UserNotFound)
	}
	resp := convert.UserToAPI(u)
	return &resp, nil
}
