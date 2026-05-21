package runner

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	imkafka "im/pkg/kafka"
	"im/pkg/models"
)

type InboxUnread struct {
	svc         *svc.ServiceContext
	memberBatch int
	mergeWindow time.Duration
}

func NewInboxUnread(s *svc.ServiceContext) *InboxUnread {
	batch := s.Config.Cron.MemberBatch
	if batch <= 0 {
		batch = 500
	}
	ms := s.Config.Cron.InboxMergeMs
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
	reader := imkafka.NewReader(r.svc.Config.Kafka.Brokers, events.TopicMessageSend, "inbox-unread")
	defer reader.Close()
	inboxWriter := imkafka.NewWriter(r.svc.Config.Kafka.Brokers, events.TopicInboxUpdated)
	defer inboxWriter.Close()
	batcher := newInboxBatcher(inboxWriter, r.mergeWindow)

	log.Println("[cron] inbox-unread started")
	for {
		select {
		case <-ctx.Done():
			batcher.flushAll(ctx)
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
		r.process(ctx, evt, batcher)
		_ = reader.CommitMessages(ctx, m)
	}
}

func (r *InboxUnread) process(ctx context.Context, evt events.MessageSendEvent, b *inboxBatcher) {
	pool := r.svc.Pool
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
	_ = forEachGroupMember(ctx, pool, evt.GroupID, r.memberBatch, func(uid int64) {
		if uid == evt.SenderID || muted(ctx, pool, evt.GroupID, uid) {
			return
		}
		_ = rdb.IncrUnread(ctx, uid, evt.ConvID, 1)
		b.add(events.InboxUpdatedEvent{
			UserID: uid, ConvID: evt.ConvID, ConvType: evt.ConvType,
			Seq: evt.Seq, UnreadDelta: 1, Ts: evt.Ts,
		})
	})
}

func muted(ctx context.Context, pool *pgxpool.Pool, groupID, uid int64) bool {
	var m bool
	err := pool.QueryRow(ctx,
		`SELECT muted FROM group_members WHERE group_id=$1 AND user_id=$2`, groupID, uid,
	).Scan(&m)
	return err == nil && m
}

type inboxBatcher struct {
	mu      sync.Mutex
	pending map[int64]events.InboxUpdatedEvent
	writer  *kafkago.Writer
	window  time.Duration
	timer   *time.Timer
}

func newInboxBatcher(w *kafkago.Writer, window time.Duration) *inboxBatcher {
	return &inboxBatcher{
		pending: make(map[int64]events.InboxUpdatedEvent),
		writer:  w,
		window:  window,
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
		_ = imkafka.PublishJSON(ctx, b.writer, strconv.FormatInt(uid, 10), evt)
	}
	b.pending = make(map[int64]events.InboxUpdatedEvent)
}
