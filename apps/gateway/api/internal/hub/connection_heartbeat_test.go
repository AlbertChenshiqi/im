package hub

import "testing"

func TestConnectionHeartbeat_threeMissesDisconnect(t *testing.T) {
	hb := NewConnectionHeartbeat(3)

	if !hb.OnServerTick() {
		t.Fatal("tick1 should stay")
	}
	for i := 2; i <= 4; i++ {
		ok := hb.OnServerTick()
		if i < 4 && !ok {
			t.Fatalf("tick%d should stay", i)
		}
		if i == 4 && ok {
			t.Fatal("tick4 should fail after 3 consecutive misses")
		}
	}
}

func TestConnectionHeartbeat_ackResetsMiss(t *testing.T) {
	hb := NewConnectionHeartbeat(3)
	_ = hb.OnServerTick()
	_ = hb.OnServerTick()
	hb.Ack()
	if !hb.OnServerTick() {
		t.Fatal("after ack, next tick should stay")
	}
}
