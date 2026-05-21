package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/user/rpc/internal/convert"
	"im/apps/user/rpc/internal/svc"
	"im/apps/user/rpc/user"
)

type BatchGetUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUsersLogic {
	return &BatchGetUsersLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *BatchGetUsersLogic) BatchGetUsers(in *user.BatchGetUsersReq) (*user.BatchGetUsersResp, error) {
	users, err := l.svcCtx.UserRepo.BatchGetByIDs(l.ctx, in.Ids)
	if err != nil {
		return nil, err
	}
	out := make([]*user.UserInfo, 0, len(users))
	for _, u := range users {
		out = append(out, convert.UserToRPC(u))
	}
	return &user.BatchGetUsersResp{Users: out}, nil
}
