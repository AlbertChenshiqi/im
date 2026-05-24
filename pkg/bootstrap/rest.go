package bootstrap

import (
	"github.com/zeromicro/go-zero/rest"

	"im/pkg/httpware"
)

// UseRESTAccessLog 为 REST 服务注册业务访问日志（跳过 health、gateway WebSocket）。
func UseRESTAccessLog(s *rest.Server) {
	s.Use(httpware.AccessLog)
}
