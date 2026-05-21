package middleware

import (
	"net/http"
	"strings"

	"im/pkg/jwtx"
)

// AuthenticateWS 从连接参数解析 JWT（query token 或 Authorization: Bearer）
func AuthenticateWS(r *http.Request, secret string) (int64, error) {
	token := ExtractToken(r)
	return jwtx.ParseUserID(secret, token)
}

func ExtractToken(r *http.Request) string {
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return ""
}
