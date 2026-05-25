package runner

import (
	"context"
	"encoding/json"
	"log"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	"im/pkg/models"
	"im/pkg/rocketmq"
)

// RealtimeMessage 订阅 message_send，向在线用户推送消息正文（WebSocket）。
type RealtimeMessage struct {
	svc *svc.ServiceContext
}

func NewRealtimeMessage(s *svc.ServiceContext) *RealtimeMessage {
	return &RealtimeMessage{svc: s}
}

func (r *RealtimeMessage) Run(ctx context.Context) {
	ns := r.svc.Config.RocketMQ.NameServer
	log.Println("[cron] realtime-message started")
	_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
		NameServers: ns,
		Topic:       events.TopicChat,
		Group:       "realtime-message",
		Tag:         events.ChatSubscribeAll,
	}, func(ctx context.Context, body []byte) error {
		var evt events.MessageSendEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil
		}
		r.fanout(ctx, evt)
		return nil
	})
}

func (r *RealtimeMessage) fanout(ctx context.Context, evt events.MessageSendEvent) {
	targets := r.targets(ctx, evt)
	frame := events.MessageFrame(evt)
	for _, uid := range targets {
		if uid == evt.SenderID {
			continue
		}
		if !userOnline(ctx, r.svc, uid) {
			continue
		}
		if err := events.PublishGatewayPush(ctx, r.svc.GatewayPush, uid, frame); err != nil {
			log.Printf("[cron] gateway message uid=%d msg=%d: %v", uid, evt.MsgID, err)
		}
	}
}

func (r *RealtimeMessage) targets(ctx context.Context, evt events.MessageSendEvent) []int64 {
	if evt.ConvType == models.ConvTypeC2C || evt.ConvType == models.ConvTypeDirect {
		return evt.RecipientIDs
	}
	var out []int64
	err := forEachGroupMember(ctx, r.svc.DB, evt.GroupID, r.svc.Config.Cron.MemberBatch, func(uid int64) {
		if uid > 0 && uid != evt.SenderID {
			out = append(out, uid)
		}
	})
	if err != nil {
		log.Printf("[cron] realtime list members gid=%d: %v", evt.GroupID, err)
		return nil
	}
	return out
}
