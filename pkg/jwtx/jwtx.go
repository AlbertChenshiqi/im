package jwtx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(secret string, userID, expireSec int64) (string, error) {
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"exp":    now + expireSec,
		"iat":    now,
		"userId": userID,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func UserIDFromCtx(ctx context.Context) int64 {
	v := ctx.Value("userId")
	switch t := v.(type) {
	case json.Number:
		id, _ := t.Int64()
		return id
	case float64:
		return int64(t)
	case int64:
		return t
	}
	return 0
}
