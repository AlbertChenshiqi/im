package jwtx

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

// ParseUserID 校验 JWT 并解析 userId（与登录签发格式一致）
func ParseUserID(secret, token string) (int64, error) {
	if token == "" {
		return 0, ErrInvalidToken
	}
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !t.Valid {
		return 0, ErrInvalidToken
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}
	raw, ok := claims["userId"]
	if !ok {
		return 0, ErrInvalidToken
	}
	switch v := raw.(type) {
	case float64:
		uid := int64(v)
		if uid <= 0 {
			return 0, ErrInvalidToken
		}
		return uid, nil
	case int64:
		if v <= 0 {
			return 0, ErrInvalidToken
		}
		return v, nil
	default:
		return 0, ErrInvalidToken
	}
}
