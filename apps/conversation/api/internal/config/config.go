package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	MySQL struct{ DSN string }
	Redis struct {
		Addr string
	}
	// Conversation 列表策略
	Conversation struct {
		// DirectRecentDays 私信默认时间窗：0=全部，>0=仅近 N 天有活动（可被 query direct_days 覆盖）
		DirectRecentDays int `json:",default=0"`
	}
}
