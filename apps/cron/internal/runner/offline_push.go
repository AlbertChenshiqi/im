package runner

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"im/apps/cron/internal/svc"
	"im/pkg/events"
	imkafka "im/pkg/kafka"
	"im/pkg/offlinepush"
)

type OfflinePush struct {
	svc      *svc.ServiceContext
	mergeSec int
}

func NewOfflinePush(s *svc.ServiceContext) *OfflinePush {
	sec := s.Config.Cron.OfflineMergeSec
	if sec <= 0 {
		sec = 10
	}
	if v := os.Getenv("OFFLINE_MERGE_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sec = n
		}
	}
	return &OfflinePush{svc: s, mergeSec: sec}
}

func (r *OfflinePush) Run(ctx context.Context) {
	reader := imkafka.NewReader(r.svc.Config.Kafka.Brokers, events.TopicPushOffline, "offline-push")
	defer reader.Close()
	vendor := offlinepush.NewVendor()
	agg := newOfflineAggregator(r.mergeSec, vendor)

	log.Printf("[cron] offline-push started (merge=%ds)", r.mergeSec)
	for {
		select {
		case <-ctx.Done():
			agg.flushAll()
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
		var evt events.PushOfflineEvent
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			_ = reader.CommitMessages(ctx, m)
			continue
		}
		agg.add(evt)
		_ = reader.CommitMessages(ctx, m)
	}
}

type offlineAggregator struct {
	mu      sync.Mutex
	pending map[int64]*events.PushOfflineEvent
	window  time.Duration
	vendor  *offlinepush.Vendor
	timers  map[int64]*time.Timer
}

func newOfflineAggregator(mergeSec int, v *offlinepush.Vendor) *offlineAggregator {
	return &offlineAggregator{
		pending: make(map[int64]*events.PushOfflineEvent),
		window:  time.Duration(mergeSec) * time.Second,
		vendor:  v,
		timers:  make(map[int64]*time.Timer),
	}
}

func (a *offlineAggregator) add(evt events.PushOfflineEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if p, ok := a.pending[evt.UserID]; ok {
		p.Count += evt.Count
		if evt.Body != "" {
			p.Body = evt.Body
		}
	} else {
		copy := evt
		a.pending[evt.UserID] = &copy
	}
	if _, ok := a.timers[evt.UserID]; !ok {
		uid := evt.UserID
		a.timers[uid] = time.AfterFunc(a.window, func() {
			a.flush(uid)
		})
	}
}

func (a *offlineAggregator) flush(uid int64) {
	a.mu.Lock()
	evt, ok := a.pending[uid]
	delete(a.pending, uid)
	if t, ok := a.timers[uid]; ok {
		t.Stop()
		delete(a.timers, uid)
	}
	a.mu.Unlock()
	if !ok || evt == nil {
		return
	}
	body := evt.Body
	if evt.Count > 1 {
		body = "您有 " + strconv.Itoa(evt.Count) + " 条新消息"
	}
	log.Printf("[cron] offline-push uid=%d title=%s body=%s", uid, evt.Title, body)
	_ = a.vendor.Send(context.Background(), uid, evt.Title, body)
}

func (a *offlineAggregator) flushAll() {
	a.mu.Lock()
	uids := make([]int64, 0, len(a.pending))
	for uid := range a.pending {
		uids = append(uids, uid)
	}
	a.mu.Unlock()
	for _, uid := range uids {
		a.flush(uid)
	}
}
