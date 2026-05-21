package grpcware

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryAccessLog 记录 gRPC 业务调用。
func UnaryAccessLog(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	code := status.Code(err)
	if err != nil {
		logx.WithContext(ctx).Infof("[rpc] %s code=%s dur=%s err=%v", info.FullMethod, code, time.Since(start), err)
	} else {
		logx.WithContext(ctx).Infof("[rpc] %s code=%s dur=%s", info.FullMethod, code, time.Since(start))
	}
	return resp, err
}
