package config

import (
	"testing"
	"time"
)

func TestWebSocketConf_HeartbeatAndReadTimeout(t *testing.T) {
	c := WebSocketConf{HeartbeatInterval: 60}
	if got := c.HeartbeatIntervalDuration(); got != 60*time.Second {
		t.Fatalf("heartbeat interval: got %v want 60s", got)
	}
	if got := c.ReadTimeoutDuration(); got != 240*time.Second {
		t.Fatalf("read timeout: got %v want 240s (60*(3+1))", got)
	}

	c2 := WebSocketConf{HeartbeatInterval: 20}
	if got := c2.ReadTimeoutDuration(); got != 90*time.Second {
		t.Fatalf("read timeout min: got %v want 90s", got)
	}
}
