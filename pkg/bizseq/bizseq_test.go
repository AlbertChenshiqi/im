package bizseq

import "testing"

func TestCompose(t *testing.T) {
	seq, err := Compose(100, 42)
	if err != nil {
		t.Fatal(err)
	}
	if seq != (100<<SlotOffsetBits)|42 {
		t.Fatalf("got %d", seq)
	}
}

func TestTimeSlot(t *testing.T) {
	if TimeSlot(499) != 0 {
		t.Fatalf("slot=%d", TimeSlot(499))
	}
	if TimeSlot(500) != 1 {
		t.Fatalf("slot=%d", TimeSlot(500))
	}
}

func TestComposeFromRecvMs(t *testing.T) {
	seq, err := ComposeFromRecvMs(400, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := int64((400 << SlotOffsetBits) | 0)
	if seq != want {
		t.Fatalf("got %d want %d", seq, want)
	}
}
