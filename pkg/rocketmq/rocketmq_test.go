package rocketmq

import (
	"net"
	"testing"
)

func TestResolveNameServers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   []string
		want string // expected host part after resolve
	}{
		{[]string{"127.0.0.1:9876"}, "127.0.0.1"},
		{[]string{"localhost:9876"}, "127.0.0.1"},
	}
	for _, tc := range cases {
		got, err := resolveNameServers(tc.in)
		if err != nil {
			t.Fatalf("resolveNameServers(%v): %v", tc.in, err)
		}
		host, _, err := net.SplitHostPort(got[0])
		if err != nil {
			t.Fatal(err)
		}
		if host != tc.want {
			t.Fatalf("got host %q want %q (full %v)", host, tc.want, got)
		}
	}
}
