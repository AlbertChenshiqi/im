package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/conversation/api/internal/svc"
	"im/apps/conversation/api/internal/types"
	"im/pkg/jwtx"
	"im/pkg/models"
	"im/pkg/repo"
)

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListLogic) List(req *types.ListConversationsReq) (*types.ConversationsResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	directDays := l.svcCtx.Config.Conversation.DirectRecentDays
	if req != nil && req.DirectDays != nil {
		directDays = *req.DirectDays
	}

	groups, err := l.svcCtx.ConvRepo.ListGroupsForUser(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	directs, err := l.svcCtx.ConvRepo.ListDirectForUser(l.ctx, uid, directDays)
	if err != nil {
		return nil, err
	}
	list := repo.MergeConversationRows(groups, directs)

	unread, _ := l.svcCtx.Redis.GetAllUnread(l.ctx, uid)
	out := make([]types.Conversation, 0, len(list))
	for _, c := range list {
		typ := c.Type
		if typ == models.ConvTypeDirect {
			typ = models.ConvTypeC2C
		}
		out = append(out, types.Conversation{
			Id: c.ID, Type: typ, GroupId: c.GroupID, GroupName: c.GroupName,
			PeerUserId: c.PeerUserID, LastSeq: c.LastSeq, LastPreview: c.LastPreview,
			Unread: unread[c.ID], Pinned: c.Pinned, Muted: c.Muted,
		})
	}
	l.Infof("[conversation] list ok uid=%d groups=%d directs=%d total=%d", uid, len(groups), len(directs), len(out))
	return &types.ConversationsResp{Conversations: out}, nil
}
