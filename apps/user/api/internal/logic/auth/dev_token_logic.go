package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/api/internal/convert"
	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
	"im/pkg/jwtx"
)

type DevTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDevTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DevTokenLogic {
	return &DevTokenLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DevTokenLogic) DevToken(req *types.DevTokenReq) (*types.DevTokenResp, error) {
	if !l.svcCtx.Config.Auth.DevMode {
		return nil, code.New(code.UserDevAuthDisabled)
	}
	if req.UserId <= 0 {
		return nil, code.New(code.UserIDRequired)
	}
	u, err := l.svcCtx.UserRepo.EnsureDevUser(l.ctx, req.UserId)
	if err != nil {
		return nil, code.New(code.UserRegisterFailed, err.Error())
	}
	token, err := jwtx.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, u.ID, l.svcCtx.Config.Auth.AccessExpire)
	if err != nil {
		return nil, code.New(code.CommonInternal, err.Error())
	}
	l.Infof("[user] dev-token ok uid=%d", u.ID)
	return &types.DevTokenResp{
		Token:  token,
		UserId: u.ID,
		User:   convert.UserToAPI(u),
	}, nil
}
