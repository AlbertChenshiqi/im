package bizseq

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// SlotDivisorMs 时间分片粒度（200ms 一片）。
	SlotDivisorMs = 200
	SlotOffsetBits = 24
	maxSlotOffset  = (1 << SlotOffsetBits) - 1
	slotKeyTTL     = 24 * time.Hour
)

// TimeSlot 服务端接收时间戳(ms) → 时间片。
func TimeSlot(serverRecvMs int64) int64 {
	return serverRecvMs / SlotDivisorMs
}

// SlotKey Redis 分片序号 Key：im:seq:slot:{sessionId}:{timeSlot}
func SlotKey(sessionID string, slot int64) string {
	return fmt.Sprintf("im:seq:slot:%s:%d", sessionID, slot)
}

// Compose 拼接 bizSeq = slot<<24 | slotOffset
func Compose(slot, slotOffset int64) (int64, error) {
	if slotOffset < 0 || slotOffset > maxSlotOffset {
		return 0, fmt.Errorf("bizseq slot offset overflow: %d", slotOffset)
	}
	return (slot << SlotOffsetBits) | slotOffset, nil
}

// Allocate 在同一时间片内 INCR 得到偏移并合成 bizSeq。
func Allocate(ctx context.Context, rdb *redis.Client, sessionID string, serverRecvMs int64) (int64, error) {
	slot := TimeSlot(serverRecvMs)
	key := SlotKey(sessionID, slot)
	offset, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if offset == 1 {
		_ = rdb.Expire(ctx, key, slotKeyTTL).Err()
	}
	return Compose(slot, offset)
}
