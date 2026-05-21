package runner

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultMemberBatch = 500

// forEachGroupMember 遍历群成员（分页），用于全局在线推送与未读扇出。
func forEachGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID int64, batch int, fn func(uid int64)) error {
	if batch <= 0 {
		batch = defaultMemberBatch
	}
	var cursor int64
	for {
		ids, next, err := listMembers(ctx, pool, groupID, cursor, batch)
		if err != nil {
			return err
		}
		for _, uid := range ids {
			fn(uid)
		}
		if len(ids) < batch {
			return nil
		}
		cursor = next
	}
}
