package hub

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

// RunConnectionHeartbeat 在连接存活期间定时续期 Redis online，并经 writePump 发送 WS 协议 Ping。
// 连续 maxMiss 次未收到心跳响应（协议 Pong / 应用层 ping）则关闭连接。
// ctx 在读循环结束时 cancel。
func (h *Hub) RunConnectionHeartbeat(ctx context.Context, uid int64, conn *websocket.Conn, client *Client, hb *ConnectionHeartbeat) {
	interval := h.WsConfig().HeartbeatIntervalDuration()
	if interval <= 0 {
		return
	}
	readTimeout := h.WsConfig().ReadTimeoutDuration()

	refreshRead := func() {
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
	}
	refreshRead()

	conn.SetPongHandler(func(string) error {
		refreshRead()
		hb.Ack()
		h.TouchOnline(ctx, uid)
		return nil
	})

	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if !hb.OnServerTick() {
					logx.Infof("[gateway] ws heartbeat timeout uid=%d missed=%d, closing",
						uid, h.WsConfig().HeartbeatMaxMissCount())
					client.Close()
					return
				}
				h.TouchOnline(ctx, uid)
				if !client.EnqueuePing() {
					return
				}
			}
		}
	}()
}

// RefreshReadDeadline 在收到任意上行帧后调用，延长读超时（心跳 miss 仍由 ConnectionHeartbeat 统计）。
func (h *Hub) RefreshReadDeadline(conn *websocket.Conn) {
	interval := h.WsConfig().HeartbeatIntervalDuration()
	if interval <= 0 {
		return
	}
	_ = conn.SetReadDeadline(time.Now().Add(h.WsConfig().ReadTimeoutDuration()))
}
