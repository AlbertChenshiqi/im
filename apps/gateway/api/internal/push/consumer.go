package push

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/gateway/api/internal/hub"
	"github.com/segmentio/kafka-go"

	"im/pkg/events"
	imkafka "im/pkg/kafka"
)

// Consumer Kafka im.gateway.push → 本机 WebSocket 广播
type Consumer struct {
	reader  *kafka.Reader
	hub     *hub.Hub
	groupID string
}

func NewConsumer(brokers []string, h *hub.Hub) *Consumer {
	groupID := os.Getenv("GATEWAY_INSTANCE_ID")
	if groupID == "" {
		host, _ := os.Hostname()
		groupID = "gateway-push-" + host
	}
	return &Consumer{
		reader:  imkafka.NewReader(brokers, events.TopicGatewayPush, groupID),
		hub:     h,
		groupID: groupID,
	}
}

// Start 阻塞消费直到 ctx 取消
func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer c.reader.Close()
		logx.Infof("[gateway] push consumer started topic=%s group=%s instance=%s",
			events.TopicGatewayPush, c.groupID, c.hub.InstanceID())
		for {
			select {
			case <-ctx.Done():
				logx.Info("[gateway] push consumer stopping")
				return
			default:
			}
			m, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				time.Sleep(time.Second)
				continue
			}
			c.dispatch(m.Value)
			_ = c.reader.CommitMessages(ctx, m)
		}
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
