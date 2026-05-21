package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/friend/api/internal/svc"
	"im/apps/friend/api/internal/types"
	"im/pkg/jwtx"
)

type RequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestLogic {
	return &RequestLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RequestLogic) Request(req *types.FriendReq) (*types.AcceptResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	if err := l.svcCtx.FriendRepo.CreateRequest(l.ctx, uid, req.UserId); err != nil {
		return nil, err
	}
	return &types.AcceptResp{Status: "pending"}, nil
}
