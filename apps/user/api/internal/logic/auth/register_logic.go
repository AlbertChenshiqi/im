package auth

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	"im/pkg/jwtx"
	"im/apps/user/api/internal/convert"
	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.AuthResp, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	nick := req.Nickname
	if nick == "" {
		nick = req.Username
	}
	u, err := l.svcCtx.UserRepo.CreateUser(l.ctx, req.Username, string(hash), nick)
	if err != nil {
		return nil, err
	}
	token, err := jwtx.GenerateToken(l.svcCtx.Config.Auth.AccessSecret, u.ID, l.svcCtx.Config.Auth.AccessExpire)
	if err != nil {
		return nil, err
	}
	l.Infof("[user] register ok uid=%d username=%s", u.ID, req.Username)
	return &types.AuthResp{Token: token, User: convert.UserToAPI(u)}, nil
}
