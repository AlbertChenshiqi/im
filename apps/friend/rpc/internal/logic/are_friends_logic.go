package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/friend/rpc/friend"
	"im/apps/friend/rpc/internal/svc"
)

type AreFriendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAreFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AreFriendsLogic {
	return &AreFriendsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AreFriendsLogic) AreFriends(in *friend.AreFriendsReq) (*friend.AreFriendsResp, error) {
	ok, _ := l.svcCtx.FriendRepo.AreFriends(l.ctx, in.UserA, in.UserB)
	return &friend.AreFriendsResp{Ok: ok}, nil
}
