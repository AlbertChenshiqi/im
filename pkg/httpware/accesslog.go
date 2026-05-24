package httpware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// AccessLog 记录 HTTP API 业务访问（不含 WebSocket 升级与健康检查）。
func AccessLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if skipAccessLog(r.URL.Path) {
			next(w, r)
			return
		}
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next(sw, r)
		logx.Infof("[api] %s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	}
}

func skipAccessLog(path string) bool {
	switch path {
	case "/health", "/gateway/v1/health", "/gateway/v1/ws":
		return true
	default:
		return false
	}
}
