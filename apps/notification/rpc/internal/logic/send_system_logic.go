package logic

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/notification/rpc/internal/svc"
	"im/apps/notification/rpc/notification"
	"im/pkg/events"
)

type SendSystemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSystemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSystemLogic {
	return &SendSystemLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SendSystemLogic) SendSystem(in *notification.SendSystemReq) (*notification.SendSystemResp, error) {
	cat := in.Category
	if cat == "" {
		cat = "system"
	}
	n, err := l.svcCtx.NotificationRepo.Create(l.ctx, in.UserId, in.Title, in.Body, cat)
	if err != nil {
		return nil, err
	}
	evt := events.NotificationEvent{UserID: in.UserId, Title: in.Title, Body: in.Body, Category: cat}
	_ = l.svcCtx.Producer.PublishJSON(l.ctx, events.TopicPush, events.TagSystemAnnounce, strconv.FormatInt(in.UserId, 10), evt)
	return &notification.SendSystemResp{Id: n.ID}, nil
}
