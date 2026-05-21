package runner

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func listMembers(ctx context.Context, pool *pgxpool.Pool, groupID, cursor int64, limit int) ([]int64, int64, error) {
	rows, err := pool.Query(ctx,
		`SELECT user_id FROM group_members WHERE group_id=$1 AND user_id > $2 ORDER BY user_id LIMIT $3`,
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
