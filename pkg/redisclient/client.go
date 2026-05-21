package redisclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	RDB *redis.Client
}

func New(addr string) *Client {
	return &Client{RDB: redis.NewClient(&redis.Options{Addr: addr})}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.RDB.Ping(ctx).Err()
}

func ConvSeqKey(convID string) string {
	return fmt.Sprintf("conv:%s:seq", convID)
}

func UnreadKey(uid int64) string {
	return fmt.Sprintf("unread:%d", uid)
}

func OnlineKey(uid int64) string {
	return fmt.Sprintf("online:%d", uid)
}

// OnlineGatewaysKey 记录哪些 gateway 实例持有该用户的 WS（多副本 gateway 用）
func OnlineGatewaysKey(uid int64) string {
	return fmt.Sprintf("online_gateways:%d", uid)
}

func DedupeKey(clientMsgID string) string {
	return fmt.Sprintf("dedupe:%s", clientMsgID)
}

func (c *Client) IncrConvSeq(ctx context.Context, convID string) (int64, error) {
	return c.RDB.Incr(ctx, ConvSeqKey(convID)).Result()
}

func (c *Client) IncrUnread(ctx context.Context, uid int64, convID string, delta int64) error {
	return c.RDB.HIncrBy(ctx, UnreadKey(uid), convID, delta).Err()
}

func (c *Client) GetUnread(ctx context.Context, uid int64, convID string) (int64, error) {
	v, err := c.RDB.HGet(ctx, UnreadKey(uid), convID).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

func (c *Client) GetAllUnread(ctx context.Context, uid int64) (map[string]int64, error) {
	m, err := c.RDB.HGetAll(ctx, UnreadKey(uid)).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(m))
	for k, v := range m {
		n, _ := strconv.ParseInt(v, 10, 64)
		out[k] = n
	}
	return out, nil
}

func (c *Client) SetOnline(ctx context.Context, uid int64, ttlSec int) error {
	if ttlSec <= 0 {
		ttlSec = 300
	}
	return c.RDB.SetEx(ctx, OnlineKey(uid), "1", time.Duration(ttlSec)*time.Second).Err()
}

func (c *Client) TouchOnline(ctx context.Context, uid int64, ttlSec int) error {
	return c.SetOnline(ctx, uid, ttlSec)
}

// AddGatewayPresence 用户在本 gateway 实例上线（多副本时聚合 online:{uid}）
func (c *Client) AddGatewayPresence(ctx context.Context, uid int64, gatewayID string, ttlSec int) error {
	if ttlSec <= 0 {
		ttlSec = 300
	}
	ttl := time.Duration(ttlSec) * time.Second
	pipe := c.RDB.Pipeline()
	pipe.SAdd(ctx, OnlineGatewaysKey(uid), gatewayID)
	pipe.Expire(ctx, OnlineGatewaysKey(uid), ttl)
	pipe.SetEx(ctx, OnlineKey(uid), "1", ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// TouchGatewayPresence 续期本实例持有的在线状态
func (c *Client) TouchGatewayPresence(ctx context.Context, uid int64, gatewayID string, ttlSec int) error {
	return c.AddGatewayPresence(ctx, uid, gatewayID, ttlSec)
}

// RemoveGatewayPresence 本实例已无该用户连接；仅当无任何 gateway 实例持有连接时删除 online:{uid}
func (c *Client) RemoveGatewayPresence(ctx context.Context, uid int64, gatewayID string) error {
	if err := c.RDB.SRem(ctx, OnlineGatewaysKey(uid), gatewayID).Err(); err != nil {
		return err
	}
	n, err := c.RDB.SCard(ctx, OnlineGatewaysKey(uid)).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return c.SetOffline(ctx, uid)
	}
	return nil
}

func (c *Client) IsOnline(ctx context.Context, uid int64) (bool, error) {
	n, err := c.RDB.Exists(ctx, OnlineKey(uid)).Result()
	return n > 0, err
}

func (c *Client) SetOffline(ctx context.Context, uid int64) error {
	return c.RDB.Del(ctx, OnlineKey(uid)).Err()
}

func (c *Client) CheckDedupe(ctx context.Context, clientMsgID string, msgID int64) (bool, error) {
	if clientMsgID == "" {
		return false, nil
	}
	ok, err := c.RDB.SetNX(ctx, DedupeKey(clientMsgID), msgID, 24*time.Hour).Result()
	if err != nil {
		return false, err
	}
	return !ok, nil
}
