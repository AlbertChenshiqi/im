package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/conversation/rpc/conversation"
	"im/apps/friend/api/internal/svc"
	"im/apps/friend/api/internal/types"
	"im/pkg/jwtx"
)

type AcceptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAcceptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptLogic {
	return &AcceptLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AcceptLogic) Accept(req *types.FriendReq) (*types.AcceptResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	if err := l.svcCtx.FriendRepo.AcceptRequest(l.ctx, req.UserId, uid); err != nil {
		return nil, err
	}
	cid := ""
	if l.svcCtx.ConversationRpc != nil {
		resp, err := l.svcCtx.ConversationRpc.EnsureDirect(l.ctx, &conversation.EnsureDirectReq{UserA: uid, UserB: req.UserId})
		if err == nil && resp != nil {
			cid = resp.ConvId
		}
	}
	return &types.AcceptResp{Status: "accepted", ConvId: cid}, nil
}
