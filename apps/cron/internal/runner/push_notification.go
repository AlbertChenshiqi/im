package runner

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	"im/pkg/rocketmq"
)

// PushNotification 订阅 notification_system：在线 WebSocket，离线 push_offline。
type PushNotification struct {
	svc *svc.ServiceContext
}

func NewPushNotification(s *svc.ServiceContext) *PushNotification {
	return &PushNotification{svc: s}
}

func (r *PushNotification) Run(ctx context.Context) {
	ns := r.svc.Config.RocketMQ.NameServer
	log.Println("[cron] push-notification started")
	_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
		NameServers: ns,
		Topic:       events.TopicPush,
		Group:       "push-notification",
		Tag:         events.TagSystemAnnounce,
	}, func(ctx context.Context, body []byte) error {
		var evt events.NotificationEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil
		}
		r.handle(ctx, evt)
		return nil
	})
}

func (r *PushNotification) handle(ctx context.Context, evt events.NotificationEvent) {
	ts := time.Now().Unix()
	if userOnline(ctx, r.svc, evt.UserID) {
		frame := events.NotificationFrame(evt, ts)
		if err := events.PublishGatewayPush(ctx, r.svc.GatewayPush, evt.UserID, frame); err != nil {
			log.Printf("[cron] gateway notification uid=%d: %v", evt.UserID, err)
		}
		return
	}
	off := events.PushOfflineEvent{
		UserID: evt.UserID,
		Title:  evt.Title,
		Body:   evt.Body,
		Count:  1,
		Ts:     ts,
	}
	_ = r.svc.Producer.PublishJSON(ctx, events.TopicPush, events.TagPushOffline, strconv.FormatInt(evt.UserID, 10), off)
}
