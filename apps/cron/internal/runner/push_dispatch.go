package runner

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	"im/pkg/rocketmq"
)

// PushDispatch 订阅 inbox_updated：在线 → gateway WebSocket；离线 → push_offline。
type PushDispatch struct {
	svc *svc.ServiceContext
}

func NewPushDispatch(s *svc.ServiceContext) *PushDispatch {
	return &PushDispatch{svc: s}
}

func (r *PushDispatch) Run(ctx context.Context) {
	ns := r.svc.Config.RocketMQ.NameServer
	log.Println("[cron] push-dispatch started")
	_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
		NameServers: ns,
		Topic:       events.TopicSync,
		Group:       "push-dispatch",
		Tag:         events.TagSyncRead,
	}, func(ctx context.Context, body []byte) error {
		var evt events.InboxUpdatedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil
		}
		r.handle(ctx, evt)
		return nil
	})
}

func (r *PushDispatch) handle(ctx context.Context, evt events.InboxUpdatedEvent) {
	if userOnline(ctx, r.svc, evt.UserID) {
		unreadTotal, _ := r.svc.Redis.GetUnread(ctx, evt.UserID, evt.ConvID)
		frame := events.BadgeFrame(evt, unreadTotal)
		if err := events.PublishGatewayPush(ctx, r.svc.GatewayPush, evt.UserID, frame); err != nil {
			log.Printf("[cron] gateway badge uid=%d: %v", evt.UserID, err)
		}
		return
	}
	log.Printf("[cron] push-dispatch route offline uid=%d conv=%s (redis online key missing; WS 需上行帧/ping 续期)",
		evt.UserID, evt.ConvID)
	off := events.PushOfflineEvent{
		UserID: evt.UserID,
		ConvID: evt.ConvID,
		Title:  "新消息",
		Body:   "您有新的未读消息",
		Count:  int(evt.UnreadDelta),
		Ts:     evt.Ts,
	}
	_ = r.svc.Producer.PublishJSON(ctx, events.TopicPush, events.TagPushOffline, strconv.FormatInt(evt.UserID, 10), off)
}
