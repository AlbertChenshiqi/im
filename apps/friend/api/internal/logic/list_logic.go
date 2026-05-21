package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/friend/api/internal/svc"
	"im/apps/friend/api/internal/types"
	"im/pkg/jwtx"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListLogic) List() (*types.FriendsResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	list, err := l.svcCtx.FriendRepo.ListFriends(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]types.User, 0, len(list))
	for _, u := range list {
		out = append(out, types.User{Id: u.ID, Username: u.Username, Nickname: u.Nickname, AvatarUrl: u.AvatarURL})
	}
	return &types.FriendsResp{Friends: out}, nil
}
