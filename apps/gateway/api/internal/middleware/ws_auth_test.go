package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"im/pkg/jwtx"
)

func TestExtractToken(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/gateway/v1/ws?token=abc", nil)
	if ExtractToken(r) != "abc" {
		t.Fatal("query token")
	}
	r2 := httptest.NewRequest(http.MethodGet, "/gateway/v1/ws", nil)
	r2.Header.Set("Authorization", "Bearer xyz")
	if ExtractToken(r2) != "xyz" {
		t.Fatal("bearer token")
	}
}

func TestAuthenticateWS(t *testing.T) {
	secret := "s"
	tok, _ := jwtx.GenerateToken(secret, 9, 3600)
	r := httptest.NewRequest(http.MethodGet, "/gateway/v1/ws?token="+tok, nil)
	uid, err := AuthenticateWS(r, secret)
	if err != nil || uid != 9 {
		t.Fatalf("uid=%d err=%v", uid, err)
	}
}
