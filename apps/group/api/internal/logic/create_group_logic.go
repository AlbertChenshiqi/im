package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/group/api/internal/svc"
	"im/apps/group/api/internal/types"
	"im/pkg/code"
	"im/pkg/convid"
	"im/pkg/jwtx"
	"im/pkg/repo"
)

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *CreateGroupLogic) CreateGroup(req *types.CreateGroupReq) (*types.CreateGroupResp, error) {
	uid := jwtx.UserIDFromCtx(l.ctx)
	if uid <= 0 {
		return nil, code.New(code.CommonUnauthorized)
	}

	memberIDs := dedupePositive(req.MemberIds)
	if err := l.ensureMembers(l.ctx, memberIDs); err != nil {
		return nil, err
	}
	if _, err := l.svcCtx.UserRepo.GetByID(l.ctx, uid); err != nil {
		if l.svcCtx.Config.Auth.DevMode {
			if _, err := l.svcCtx.UserRepo.EnsureDevUser(l.ctx, uid); err != nil {
				return nil, err
			}
		} else {
			return nil, code.New(code.CommonUnauthorized)
		}
	}

	g, err := l.svcCtx.GroupRepo.CreateGroup(l.ctx, req.Name, uid, memberIDs)
	if err != nil {
		if errors.Is(err, repo.ErrTooManyMembers) {
			return nil, code.New(code.GroupTooManyMembers)
		}
		var nf *repo.ErrMembersNotFound
		if errors.As(err, &nf) {
			return nil, code.New(code.GroupMembersNotFound, nf.Error())
		}
		return nil, err
	}
	return &types.CreateGroupResp{
		Group: types.Group{
			Id: g.ID, Name: g.Name, OwnerId: g.OwnerID, MaxMembers: g.MaxMembers, Notice: g.Notice,
		},
		ConvId: convid.Group(g.ID),
	}, nil
}

func (l *CreateGroupLogic) ensureMembers(ctx context.Context, memberIDs []int64) error {
	if len(memberIDs) == 0 {
		return nil
	}
	if l.svcCtx.Config.Auth.DevMode {
		for _, id := range memberIDs {
			if _, err := l.svcCtx.UserRepo.EnsureDevUser(ctx, id); err != nil {
				return err
			}
		}
		return nil
	}
	missing, err := l.svcCtx.UserRepo.FindMissingIDs(ctx, memberIDs)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return code.New(code.GroupMembersNotFound, (&repo.ErrMembersNotFound{IDs: missing}).Error())
	}
	return nil
}

func dedupePositive(ids []int64) []int64 {
	seen := make(map[int64]bool)
	var out []int64
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
