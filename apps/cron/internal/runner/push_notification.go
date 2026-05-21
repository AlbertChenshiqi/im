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

// PushNotification 消费 im.notification.system：在线 WebSocket，离线 push.offline
type PushNotification struct {
	svc *svc.ServiceContext
}

func NewPushNotification(s *svc.ServiceContext) *PushNotification {
	return &PushNotification{svc: s}
}

func (r *PushNotification) Run(ctx context.Context) {
	reader := imkafka.NewReader(r.svc.Config.Kafka.Brokers, events.TopicNotificationSystem, "push-notification")
	defer reader.Close()
	offlineWriter := imkafka.NewWriter(r.svc.Config.Kafka.Brokers, events.TopicPushOffline)
	defer offlineWriter.Close()

	log.Println("[cron] push-notification started")
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
		var evt events.NotificationEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			_ = reader.CommitMessages(ctx, m)
			continue
		}
		r.handle(ctx, evt, offlineWriter)
		_ = reader.CommitMessages(ctx, m)
	}
}

func (r *PushNotification) handle(ctx context.Context, evt events.NotificationEvent, offline *kafkago.Writer) {
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
	_ = imkafka.PublishJSON(ctx, offline, strconv.FormatInt(evt.UserID, 10), off)
}
