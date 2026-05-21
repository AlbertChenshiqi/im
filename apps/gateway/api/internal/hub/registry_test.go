package hub

import "testing"

func TestRegistryMultiBroadcast(t *testing.T) {
	r := NewRegistry(false)
	c1 := &Client{id: 1, send: make(chan outbound, 4)}
	c2 := &Client{id: 2, send: make(chan outbound, 4)}
	r.Register(100, c1)
	r.Register(100, c2)

	n := r.Broadcast(100, map[string]string{"type": "ping"})
	if n != 2 {
		t.Fatalf("broadcast want 2 got %d", n)
	}
}

func TestRegistrySingleKicksOld(t *testing.T) {
	r := NewRegistry(true)
	old := &Client{id: 1, send: make(chan outbound, 4)}
	r.Register(100, old)
	newC := &Client{id: 2, send: make(chan outbound, 4)}
	kicked := r.Register(100, newC)
	if len(kicked) != 1 || kicked[0].id != 1 {
		t.Fatalf("expected old client kicked, got %v", kicked)
	}
	if r.ConnectionCount(100) != 1 {
		t.Fatalf("expected 1 connection")
	}
}
