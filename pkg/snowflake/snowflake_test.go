package snowflake

import "testing"

func TestNextUnique(t *testing.T) {
	g := New(1)
	seen := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id := g.Next()
		if seen[id] {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = true
	}
}
