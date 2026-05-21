package runner

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	imkafka "im/pkg/kafka"
)

// PushDispatch 消费 im.inbox.updated：在线 → gateway WebSocket；离线 → im.push.offline
type PushDispatch struct {
	svc *svc.ServiceContext
}

func NewPushDispatch(s *svc.ServiceContext) *PushDispatch {
	return &PushDispatch{svc: s}
}

func (r *PushDispatch) Run(ctx context.Context) {
	reader := imkafka.NewReader(r.svc.Config.Kafka.Brokers, events.TopicInboxUpdated, "push-dispatch")
	defer reader.Close()
	offlineWriter := imkafka.NewWriter(r.svc.Config.Kafka.Brokers, events.TopicPushOffline)
	defer offlineWriter.Close()

	log.Println("[cron] push-dispatch started")
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
		var evt events.InboxUpdatedEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			_ = reader.CommitMessages(ctx, m)
			continue
		}
		r.handle(ctx, evt, offlineWriter)
		_ = reader.CommitMessages(ctx, m)
	}
}

func (r *PushDispatch) handle(ctx context.Context, evt events.InboxUpdatedEvent, offline *kafkago.Writer) {
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
	_ = imkafka.PublishJSON(ctx, offline, strconv.FormatInt(evt.UserID, 10), off)
}
