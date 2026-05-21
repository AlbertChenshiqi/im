package runner

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	imkafka "im/pkg/kafka"
	"im/pkg/models"
)

// RealtimeMessage 消费 im.message.send，向在线用户推送消息正文（WebSocket）
type RealtimeMessage struct {
	svc *svc.ServiceContext
}

func NewRealtimeMessage(s *svc.ServiceContext) *RealtimeMessage {
	return &RealtimeMessage{svc: s}
}

func (r *RealtimeMessage) Run(ctx context.Context) {
	reader := imkafka.NewReader(r.svc.Config.Kafka.Brokers, events.TopicMessageSend, "realtime-message")
	defer reader.Close()

	log.Println("[cron] realtime-message started")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(time.Second)
			continue
		}
		var evt events.MessageSendEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			_ = reader.CommitMessages(ctx, m)
			continue
		}
		r.fanout(ctx, evt)
		_ = reader.CommitMessages(ctx, m)
	}
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
	err := forEachGroupMember(ctx, r.svc.Pool, evt.GroupID, r.svc.Config.Cron.MemberBatch, func(uid int64) {
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
