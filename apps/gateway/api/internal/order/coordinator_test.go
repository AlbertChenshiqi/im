package order

import "testing"

func TestUseRedisForSeq(t *testing.T) {
	const windowMs int64 = 200

	cases := []struct {
		name         string
		lastRecvMs   int64
		serverRecvMs int64
		want         bool
	}{
		{"first message", 0, 1000, false},
		{"within window", 1000, 1150, true},
		{"exactly window", 1000, 1200, true},
		{"after window", 1000, 1201, false},
		{"gap", 5000, 6000, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := useRedisForSeq(tc.lastRecvMs, tc.serverRecvMs, windowMs)
			if got != tc.want {
				t.Fatalf("useRedisForSeq(%d,%d,%d)=%v want %v",
					tc.lastRecvMs, tc.serverRecvMs, windowMs, got, tc.want)
			}
		})
	}
}
