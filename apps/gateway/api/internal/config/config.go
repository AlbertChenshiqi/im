package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
	}
	MessageRpc struct {
		Endpoints []string
	}
	Redis struct {
		Addr string
	}
	RocketMQ struct {
		NameServer []string
	}
	WebSocket WebSocketConf
	// SendOrder 服务端统一排序：聚合窗口(ms)、与 bizseq 时间分片一致默认 200
	SendOrder SendOrderConf `json:",optional"`
}

type SendOrderConf struct {
	WindowMs int `json:",default=200"`
}

type WebSocketConf struct {
	OnlineTTL           int      `json:",default=300"`
	HeartbeatInterval   int      `json:",default=60"`  // 服务端续期 online / WS Ping 间隔（秒），应 < OnlineTTL/2
	HeartbeatMaxMiss    int      `json:",default=3"`   // 连续未响应心跳次数达到该值则断开 WS
	MaxMessageBytes     int64    `json:",default=65536"`
	ConnectionMode      string   `json:",default=multi"` // multi | single
	AllowedOrigins      []string `json:",optional"`
}

func (c WebSocketConf) IsSingleConnection() bool {
	return c.ConnectionMode == "single"
}

func (c WebSocketConf) OnlineTTLSeconds() int {
	if c.OnlineTTL <= 0 {
		return 300
	}
	return c.OnlineTTL
}

func (c WebSocketConf) MaxMessageSize() int64 {
	if c.MaxMessageBytes <= 0 {
		return 65536
	}
	return c.MaxMessageBytes
}

func (c WebSocketConf) HeartbeatIntervalSeconds() int {
	if c.HeartbeatInterval <= 0 {
		return 60
	}
	return c.HeartbeatInterval
}

func (c WebSocketConf) HeartbeatMaxMissCount() int {
	if c.HeartbeatMaxMiss <= 0 {
		return 3
	}
	return c.HeartbeatMaxMiss
}

func (c WebSocketConf) HeartbeatIntervalDuration() time.Duration {
	sec := c.HeartbeatIntervalSeconds()
	if sec <= 0 {
		return 0
	}
	return time.Duration(sec) * time.Second
}

// ReadTimeoutDuration 读超时兜底（心跳间隔 × maxMiss + 1，至少 90s）。
func (c WebSocketConf) ReadTimeoutDuration() time.Duration {
	hb := c.HeartbeatIntervalSeconds()
	if hb <= 0 {
		return 0
	}
	sec := hb * (c.HeartbeatMaxMissCount() + 1)
	if sec < 90 {
		sec = 90
	}
	return time.Duration(sec) * time.Second
}
