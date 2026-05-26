package runner

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"im/apps/transfer/internal/svc"
	"im/pkg/events"
	"im/pkg/offlinepush"
	"im/pkg/rocketmq"
)

type OfflinePush struct {
	svc      *svc.ServiceContext
	mergeSec int
}

func NewOfflinePush(s *svc.ServiceContext) *OfflinePush {
	sec := s.Config.Transfer.OfflineMergeSec
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
	ns := r.svc.Config.RocketMQ.NameServer
	vendor := offlinepush.NewVendor()
	agg := newOfflineAggregator(r.mergeSec, vendor)

	log.Printf("[transfer] offline-push started (merge=%ds)", r.mergeSec)
	_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
		NameServers: ns,
		Topic:       events.TopicPush,
		Group:       "offline-push",
		Tag:         events.TagPushOffline,
	}, func(ctx context.Context, body []byte) error {
		var evt events.PushOfflineEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil
		}
		agg.add(evt)
		return nil
	})
	agg.flushAll()
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
	log.Printf("[transfer] offline-push uid=%d title=%s body=%s", uid, evt.Title, body)
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
