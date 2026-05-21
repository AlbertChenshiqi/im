package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/notification/api/internal/svc"
	"im/apps/notification/api/internal/types"
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

func (l *ListLogic) List() (*types.NotificationsResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	list, err := l.svcCtx.NotificationRepo.List(l.ctx, uid, 50)
	if err != nil {
		return nil, err
	}
	out := make([]types.Notification, 0, len(list))
	for _, n := range list {
		out = append(out, types.Notification{
			Id: n.ID, Title: n.Title, Body: n.Body, Category: n.Category, Read: n.Read,
		})
	}
	return &types.NotificationsResp{Notifications: out}, nil
}
