package code

import "testing"

func TestSegment(t *testing.T) {
	if Segment(GatewayNotAuthed) != "gateway" {
		t.Fatalf("segment=%s", Segment(GatewayNotAuthed))
	}
	if Segment(UserDevAuthDisabled) != "user" {
		t.Fatalf("segment=%s", Segment(UserDevAuthDisabled))
	}
}

func TestSlugBackwardCompat(t *testing.T) {
	if GatewayNotAuthed.Slug() != "not_authed" {
		t.Fatalf("slug=%s", GatewayNotAuthed.Slug())
	}
}

func TestInRange(t *testing.T) {
	if !InRange(GatewayUnauthorized, RangeGatewayMin, RangeGatewayMax) {
		t.Fatal("gateway range")
	}
}
