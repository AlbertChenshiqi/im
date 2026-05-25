package runner

import (
	"context"
	"database/sql"
)

func listMembers(ctx context.Context, db *sql.DB, groupID, cursor int64, limit int) ([]int64, int64, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT user_id FROM group_members WHERE group_id=? AND user_id > ? ORDER BY user_id LIMIT ?`,
		groupID, cursor, limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var ids []int64
	var last int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
		last = id
	}
	return ids, last, rows.Err()
}
