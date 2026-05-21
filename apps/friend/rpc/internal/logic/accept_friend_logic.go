package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/conversation/rpc/conversation"
	"im/apps/friend/rpc/friend"
	"im/apps/friend/rpc/internal/svc"
)

type AcceptFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptFriendLogic {
	return &AcceptFriendLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *AcceptFriendLogic) AcceptFriend(in *friend.AcceptFriendReq) (*friend.AcceptFriendResp, error) {
	if err := l.svcCtx.FriendRepo.AcceptRequest(l.ctx, in.FromUserId, in.ToUserId); err != nil {
		return nil, err
	}
	cid := ""
	if l.svcCtx.ConversationRpc != nil {
		resp, err := l.svcCtx.ConversationRpc.EnsureDirect(l.ctx, &conversation.EnsureDirectReq{UserA: in.FromUserId, UserB: in.ToUserId})
		if err == nil && resp != nil {
			cid = resp.ConvId
		}
	}
	return &friend.AcceptFriendResp{ConvId: cid}, nil
}
