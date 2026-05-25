package msghandler

import (
	"testing"
)

func TestRegistryProcessText(t *testing.T) {
	reg := DefaultRegistry()
	got, err := reg.Process("text", `{"text":"hello"}`)
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"text":"hello"}` {
		t.Fatalf("got %q", got)
	}
}

func TestRegistryUnsupported(t *testing.T) {
	reg := DefaultRegistry()
	_, err := reg.Process("video", `{}`)
	if err == nil {
		t.Fatal("expected error")
	}
}
