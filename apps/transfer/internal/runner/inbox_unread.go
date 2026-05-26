package runner

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"database/sql"

	"im/apps/transfer/internal/svc"
	"im/pkg/events"
	"im/pkg/models"
	"im/pkg/rocketmq"
)

type InboxUnread struct {
	svc         *svc.ServiceContext
	memberBatch int
	mergeWindow time.Duration
}

func NewInboxUnread(s *svc.ServiceContext) *InboxUnread {
	batch := s.Config.Transfer.MemberBatch
	if batch <= 0 {
		batch = 500
	}
	ms := s.Config.Transfer.InboxMergeMs
	if ms <= 0 {
		ms = 100
	}
	return &InboxUnread{
		svc:         s,
		memberBatch: batch,
		mergeWindow: time.Duration(ms) * time.Millisecond,
	}
}

func (r *InboxUnread) Run(ctx context.Context) {
	ns := r.svc.Config.RocketMQ.NameServer
	batcher := newInboxBatcher(r.svc.Producer, r.mergeWindow)
	log.Println("[transfer] inbox-unread started")
	_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
		NameServers: ns,
		Topic:       events.TopicChat,
		Group:       "inbox-unread",
		Tag:         events.ChatSubscribeAll,
	}, func(ctx context.Context, body []byte) error {
		var evt events.MessageSendEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil
		}
		r.process(ctx, evt, batcher)
		return nil
	})
	batcher.flushAll(context.Background())
}

func (r *InboxUnread) process(ctx context.Context, evt events.MessageSendEvent, b *inboxBatcher) {
	db := r.svc.DB
	rdb := r.svc.Redis
	if evt.ConvType == models.ConvTypeC2C || evt.ConvType == models.ConvTypeDirect {
		for _, uid := range evt.RecipientIDs {
			if uid == evt.SenderID {
				continue
			}
			_ = rdb.IncrUnread(ctx, uid, evt.ConvID, 1)
			b.add(events.InboxUpdatedEvent{
				UserID: uid, ConvID: evt.ConvID, ConvType: evt.ConvType,
				Seq: evt.Seq, UnreadDelta: 1, Ts: evt.Ts,
			})
		}
		return
	}
	_ = forEachGroupMember(ctx, db, evt.GroupID, r.memberBatch, func(uid int64) {
		if uid == evt.SenderID || muted(ctx, db, evt.GroupID, uid) {
			return
		}
		_ = rdb.IncrUnread(ctx, uid, evt.ConvID, 1)
		b.add(events.InboxUpdatedEvent{
			UserID: uid, ConvID: evt.ConvID, ConvType: evt.ConvType,
			Seq: evt.Seq, UnreadDelta: 1, Ts: evt.Ts,
		})
	})
}

func muted(ctx context.Context, db *sql.DB, groupID, uid int64) bool {
	var m bool
	err := db.QueryRowContext(ctx,
		`SELECT muted FROM group_members WHERE group_id=? AND user_id=?`, groupID, uid,
	).Scan(&m)
	return err == nil && m
}

type inboxBatcher struct {
	mu       sync.Mutex
	pending  map[int64]events.InboxUpdatedEvent
	producer *rocketmq.Producer
	window   time.Duration
	timer    *time.Timer
}

func newInboxBatcher(p *rocketmq.Producer, window time.Duration) *inboxBatcher {
	return &inboxBatcher{
		pending:  make(map[int64]events.InboxUpdatedEvent),
		producer: p,
		window:   window,
	}
}

func (b *inboxBatcher) add(evt events.InboxUpdatedEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if old, ok := b.pending[evt.UserID]; ok {
		evt.UnreadDelta += old.UnreadDelta
		if evt.Seq < old.Seq {
			evt.Seq = old.Seq
		}
	}
	b.pending[evt.UserID] = evt
	if b.timer == nil {
		b.timer = time.AfterFunc(b.window, func() {
			b.flushAll(context.Background())
		})
	}
}

func (b *inboxBatcher) flushAll(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	for uid, evt := range b.pending {
		_ = b.producer.PublishJSON(ctx, events.TopicSync, events.TagSyncRead, strconv.FormatInt(uid, 10), evt)
	}
	b.pending = make(map[int64]events.InboxUpdatedEvent)
}
