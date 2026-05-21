package protocol

import "testing"

func TestInFrameValidate(t *testing.T) {
	tests := []struct {
		name    string
		frame   InFrame
		wantErr bool
	}{
		{"send ok", InFrame{Type: TypeSend, ConvId: "c2c_1_2", Content: "hi"}, false},
		{"send missing conv", InFrame{Type: TypeSend, Content: "hi"}, true},
		{"unknown", InFrame{Type: "foo"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.frame.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
