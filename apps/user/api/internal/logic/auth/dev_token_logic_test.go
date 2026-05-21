package auth

import (
	"context"
	"testing"

	"im/apps/user/api/internal/config"
	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
	"im/pkg/jwtx"
)

func TestDevTokenDisabled(t *testing.T) {
	var c config.Config
	c.Auth.DevMode = false
	l := NewDevTokenLogic(context.Background(), &svc.ServiceContext{Config: c})
	_, err := l.DevToken(&types.DevTokenReq{UserId: 1})
	if !code.Is(err, code.UserDevAuthDisabled) {
		t.Fatalf("expected UserDevAuthDisabled, got %v", err)
	}
}

func TestDevTokenGenerateJWT(t *testing.T) {
	secret := "test"
	expire := int64(3600)
	// 无 DB 时仅测 JWT 生成路径需集成测试；此处测 jwtx 与配置
	tok, err := jwtx.GenerateToken(secret, 99, expire)
	if err != nil {
		t.Fatal(err)
	}
	uid, err := jwtx.ParseUserID(secret, tok)
	if err != nil || uid != 99 {
		t.Fatalf("parse uid=%d err=%v", uid, err)
	}
}
