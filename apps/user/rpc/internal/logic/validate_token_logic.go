package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/rpc/internal/svc"
	"im/apps/user/rpc/user"
	"im/pkg/jwtx"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ValidateTokenLogic) ValidateToken(in *user.ValidateTokenReq) (*user.ValidateTokenResp, error) {
	uid, err := jwtx.ParseUserID(l.svcCtx.Config.JwtAuth.AccessSecret, in.Token)
	if err != nil {
		return &user.ValidateTokenResp{Ok: false}, nil
	}
	return &user.ValidateTokenResp{UserId: uid, Ok: true}, nil
}
