package runner

import (
	"context"
	"encoding/json"
	"log"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	"im/pkg/models"
	"im/pkg/msgbody"
	"im/pkg/rocketmq"
	"im/pkg/store"
)

type MessagePersist struct {
	svc *svc.ServiceContext
}

func NewMessagePersist(s *svc.ServiceContext) *MessagePersist {
	return &MessagePersist{svc: s}
}

func (r *MessagePersist) Run(ctx context.Context) {
	ns := r.svc.Config.RocketMQ.NameServer
	log.Println("[cron] message-persist started (topic=im_chat_persist)")
	_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
		NameServers: ns,
		Topic:       events.TopicChatPersist,
		Group:       "message-persist",
		Tag:         events.TagChatPersistStore,
	}, func(ctx context.Context, body []byte) error {
		var evt events.MessageSendEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil
		}
		seq := evt.BizSeq
		if seq <= 0 {
			seq = evt.Seq
		}
		msg := &models.Message{
			ID: evt.MsgID, ConvID: evt.ConvID, SenderID: evt.SenderID, Seq: seq,
			ClientMsgID: evt.ClientMsgID, Input: evt.Input,
			CreatedAt: store.MessageTime(evt.Ts),
		}
		if err := store.InsertMessage(ctx, r.svc.DB, msg); err != nil {
			log.Printf("[cron] insert msg conv=%s biz_seq=%d: %v", evt.ConvID, seq, err)
			return err
		}
		preview := msgbody.Preview(evt.Input)
		if len(preview) > 120 {
			preview = preview[:120]
		}
		if err := store.UpdateConvMeta(ctx, r.svc.DB, evt.ConvID, seq, evt.MsgID, preview); err != nil {
			log.Printf("[cron] update conv meta conv=%s: %v", evt.ConvID, err)
			return err
		}
		return r.svc.Producer.PublishJSON(ctx, events.TopicChat, events.TagChatPersisted, evt.SessionID, evt)
	})
}
