package convid

import "testing"

func TestDirectOrdering(t *testing.T) {
	a := C2C(5, 10)
	b := C2C(10, 5)
	if a != b {
		t.Fatalf("expected same conv id, got %s vs %s", a, b)
	}
}

func TestGroup(t *testing.T) {
	if Group(99) != "group_99" {
		t.Fatal("bad group conv id")
	}
}

func TestParseGroupID(t *testing.T) {
	id, ok := ParseGroupID("group_5")
	if !ok || id != 5 {
		t.Fatalf("parse group_5: id=%d ok=%v", id, ok)
	}
	id, ok = ParseGroupID("g_5")
	if !ok || id != 5 {
		t.Fatalf("legacy g_5: id=%d ok=%v", id, ok)
	}
	if _, ok := ParseGroupID("c2c_1_2"); ok {
		t.Fatal("c2c should not parse as group")
	}
}

func TestC2CPeer(t *testing.T) {
	peer, ok := C2CPeer("c2c_1_5", 1)
	if !ok || peer != 5 {
		t.Fatalf("peer=%d ok=%v", peer, ok)
	}
	peer, ok = C2CPeer("d_1_5", 5)
	if !ok || peer != 1 {
		t.Fatalf("legacy peer=%d ok=%v", peer, ok)
	}
}
