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
	"im/pkg/store"
)

type MessagePersist struct {
	svc *svc.ServiceContext
}

func NewMessagePersist(s *svc.ServiceContext) *MessagePersist {
	return &MessagePersist{svc: s}
}

func (r *MessagePersist) Run(ctx context.Context) {
	reader := imkafka.NewReader(r.svc.Config.Kafka.Brokers, events.TopicMessageSend, "message-persist")
	defer reader.Close()
	persisted := imkafka.NewWriter(r.svc.Config.Kafka.Brokers, events.TopicMessagePersisted)
	defer persisted.Close()

	log.Println("[cron] message-persist started")
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
		msg := &models.Message{
			ID: evt.MsgID, ConvID: evt.ConvID, SenderID: evt.SenderID, Seq: evt.Seq,
			ClientMsgID: evt.ClientMsgID, MsgType: evt.MsgType, Content: evt.Content,
			CreatedAt: store.MessageTime(evt.Ts),
		}
		if err := store.InsertMessage(ctx, r.svc.Pool, msg); err != nil {
			log.Printf("[cron] insert msg conv=%s seq=%d: %v", evt.ConvID, evt.Seq, err)
			continue
		}
		preview := evt.Content
		if len(preview) > 120 {
			preview = preview[:120]
		}
		if err := store.UpdateConvMeta(ctx, r.svc.Pool, evt.ConvID, evt.Seq, evt.MsgID, preview); err != nil {
			log.Printf("[cron] update conv meta conv=%s: %v", evt.ConvID, err)
			continue
		}
		_ = imkafka.PublishJSON(ctx, persisted, evt.ConvID, evt)
		_ = reader.CommitMessages(ctx, m)
	}
}
