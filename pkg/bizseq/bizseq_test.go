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
	if TimeSlot(399) != 1 {
		t.Fatalf("slot=%d", TimeSlot(399))
	}
	if TimeSlot(400) != 2 {
		t.Fatalf("slot=%d", TimeSlot(400))
	}
}
