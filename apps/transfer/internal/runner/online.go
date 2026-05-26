package runner

import (
	"context"
	"log"

	"im/apps/transfer/internal/svc"
)

// userOnline 以 Redis online:{uid} 为准（Gateway WS 连接/上行帧续期）。
func userOnline(ctx context.Context, s *svc.ServiceContext, uid int64) bool {
	online, err := s.Redis.IsOnline(ctx, uid)
	if err != nil {
		log.Printf("[transfer] IsOnline uid=%d err=%v", uid, err)
		return false
	}
	return online
}
