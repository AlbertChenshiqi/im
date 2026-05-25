package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	MySQL struct{ DSN string }
	ConversationRpc struct {
		Endpoints []string
	}
}
