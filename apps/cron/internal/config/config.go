package config

type Config struct {
	Name string `json:",optional"`
	// HealthPort 可选 HTTP 健康检查端口（默认 10800）
	HealthPort int `json:",optional"`
	Postgres   struct {
		DSN string
	}
	Redis struct {
		Addr string
	}
	RocketMQ struct {
		NameServer []string
	}
	Cron struct {
		InboxMergeMs    int `json:",optional"`
		OfflineMergeSec int `json:",optional"`
		MemberBatch     int `json:",optional"`
	}
}
