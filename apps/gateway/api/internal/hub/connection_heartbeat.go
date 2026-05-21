package hub

import "sync"

// ConnectionHeartbeat 跟踪服务端 WS Ping 是否得到响应（协议 Pong 或应用层 ping）。
type ConnectionHeartbeat struct {
	mu       sync.Mutex
	maxMiss  int
	missed   int
	answered bool // 上一轮 Ping 是否已收到响应
}

func NewConnectionHeartbeat(maxMiss int) *ConnectionHeartbeat {
	if maxMiss <= 0 {
		maxMiss = 3
	}
	return &ConnectionHeartbeat{
		maxMiss:  maxMiss,
		answered: true, // 连接初期无待响应 Ping，首轮不计 miss
	}
}

// Ack 收到心跳响应（协议 Pong / 应用层 ping 等）。
func (c *ConnectionHeartbeat) Ack() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.answered = true
	c.missed = 0
}

// OnServerTick 服务端定时 Ping 前调用；连续 maxMiss 次无响应返回 false，应断开连接。
func (c *ConnectionHeartbeat) OnServerTick() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.answered {
		c.missed++
		if c.missed >= c.maxMiss {
			return false
		}
	}
	c.answered = false
	return true
}
