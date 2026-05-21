package jwtx

import "testing"

func TestParseUserID(t *testing.T) {
	secret := "test-secret"
	tok, err := GenerateToken(secret, 42, 3600)
	if err != nil {
		t.Fatal(err)
	}
	uid, err := ParseUserID(secret, tok)
	if err != nil || uid != 42 {
		t.Fatalf("ParseUserID() uid=%d err=%v", uid, err)
	}
	if _, err := ParseUserID(secret, "bad"); err == nil {
		t.Fatal("expected error for bad token")
	}
}
