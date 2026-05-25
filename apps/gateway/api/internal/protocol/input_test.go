package protocol

import (
	"testing"

	"im/pkg/msghandler"
)

func TestNormalizeInput(t *testing.T) {
	reg := msghandler.DefaultRegistry()
	f := &InFrame{
		Input: []SendInputItem{{
			MsgType: "text",
			Content: `{"text":"test content"}`,
		}},
	}
	if err := NormalizeInput(f, reg); err != nil {
		t.Fatal(err)
	}
	if len(f.Input) != 1 || f.Input[0].Content != `{"text":"test content"}` {
		t.Fatalf("got %+v", f.Input)
	}
}

func TestNormalizeInputMulti(t *testing.T) {
	reg := msghandler.DefaultRegistry()
	f := &InFrame{
		Input: []SendInputItem{
			{MsgType: "image", Content: `{"url":"https://a/img.png"}`},
			{MsgType: "text", Content: `{"text":"hi"}`},
		},
	}
	if err := NormalizeInput(f, reg); err != nil {
		t.Fatal(err)
	}
	if len(f.Input) != 2 {
		t.Fatalf("len=%d", len(f.Input))
	}
}
