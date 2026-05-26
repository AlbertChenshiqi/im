// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
		DevMode      bool `json:",optional"`
	}
	MySQL struct {
		DSN string
	}
	Redis struct {
		Addr string
	}
	// OnlineTTLSeconds HTTP 心跳续期 online:{uid}（默认 300）；WebSocket 仍由 gateway 写入
	OnlineTTLSeconds int `json:",optional"`
}
