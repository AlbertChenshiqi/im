package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/conversation/api/internal/svc"
	"im/apps/conversation/api/internal/types"
	"im/pkg/jwtx"
	"im/pkg/redisclient"
)

type MarkReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *MarkReadLogic) MarkRead(req *types.MarkReadPathReq) (*types.StatusResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	_ = l.svcCtx.ConvRepo.MarkRead(l.ctx, uid, req.Id, req.Seq)
	_ = l.svcCtx.Redis.RDB.HSet(l.ctx, redisclient.UnreadKey(uid), req.Id, 0)
	return &types.StatusResp{Status: "ok"}, nil
}
