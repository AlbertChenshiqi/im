package runner

import (
	"context"
	"database/sql"
)

const defaultMemberBatch = 500

func forEachGroupMember(ctx context.Context, db *sql.DB, groupID int64, batch int, fn func(uid int64)) error {
	if batch <= 0 {
		batch = defaultMemberBatch
	}
	var cursor int64
	for {
		ids, next, err := listMembers(ctx, db, groupID, cursor, batch)
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
