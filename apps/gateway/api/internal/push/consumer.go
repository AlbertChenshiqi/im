package push

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/gateway/api/internal/hub"
	"im/pkg/events"
	"im/pkg/rocketmq"
)

// Consumer RocketMQ im / gateway_push（广播）→ 本机 WebSocket。
type Consumer struct {
	nameServers []string
	hub         *hub.Hub
	groupID     string
}

func NewConsumer(nameServers []string, h *hub.Hub) *Consumer {
	groupID := os.Getenv("GATEWAY_INSTANCE_ID")
	if groupID == "" {
		host, _ := os.Hostname()
		groupID = "gateway-push-" + host
	}
	return &Consumer{nameServers: nameServers, hub: h, groupID: groupID}
}

// Start 阻塞消费直到 ctx 取消。
func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		logx.Infof("[gateway] push consumer started topic=%s tag=%s group=%s instance=%s broadcast=true",
			events.TopicSync, events.TagSyncGateway, c.groupID, c.hub.InstanceID())
		_ = rocketmq.RunPushConsumer(ctx, rocketmq.ConsumerConfig{
			NameServers: c.nameServers,
			Topic:       events.TopicSync,
			Group:       c.groupID,
			Tag:         events.TagSyncGateway,
			Broadcast:   true,
		}, func(ctx context.Context, body []byte) error {
			c.dispatch(body)
			return nil
		})
		logx.Info("[gateway] push consumer stopping")
	}()
}

func (c *Consumer) dispatch(raw []byte) {
	var wire struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil || wire.UserID <= 0 {
		return
	}

	payload, err := decodePushPayload(raw)
	if err != nil {
		log.Printf("[gateway] push decode uid=%d: %v", wire.UserID, err)
		return
	}

	n := c.hub.Broadcast(wire.UserID, payload)
	if n == 0 {
		logx.Debugf("[gateway] push miss uid=%d (no local connection)", wire.UserID)
	} else {
		logx.Debugf("[gateway] push delivered uid=%d conns=%d", wire.UserID, n)
	}
}

func decodePushPayload(raw []byte) (any, error) {
	var peek struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &peek); err != nil {
		return nil, err
	}
	switch peek.Type {
	case "message":
		var m events.WSMessagePush
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		return m, nil
	case "badge":
		var b events.WSBadgePush
		if err := json.Unmarshal(raw, &b); err != nil {
			return nil, err
		}
		return b, nil
	case "notification":
		var n events.WSNotificationPush
		if err := json.Unmarshal(raw, &n); err != nil {
			return nil, err
		}
		return n, nil
	default:
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		delete(m, "user_id")
		return m, nil
	}
}
