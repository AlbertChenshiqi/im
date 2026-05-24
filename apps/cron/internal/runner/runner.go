package runner

import (
	"context"
	"sync"

	"im/apps/cron/internal/svc"
)

// StartAll 启动全部 RocketMQ 异步任务
func StartAll(ctx context.Context, s *svc.ServiceContext, wg *sync.WaitGroup) {
	tasks := []struct {
		name string
		run  func(context.Context)
	}{
		{"message-persist", NewMessagePersist(s).Run},
		{"inbox-unread", NewInboxUnread(s).Run},
		{"realtime-message", NewRealtimeMessage(s).Run},
		{"push-dispatch", NewPushDispatch(s).Run},
		{"offline-push", NewOfflinePush(s).Run},
		{"push-notification", NewPushNotification(s).Run},
	}
	for _, t := range tasks {
		wg.Add(1)
		go func(name string, fn func(context.Context)) {
			defer wg.Done()
			fn(ctx)
		}(t.name, t.run)
	}
}
