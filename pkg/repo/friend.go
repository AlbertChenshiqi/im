package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/models"
)

type FriendRepo struct {
	pool *pgxpool.Pool
}

func NewFriendRepo(pool *pgxpool.Pool) *FriendRepo {
	return &FriendRepo{pool: pool}
}

func (s *FriendRepo) CreateRequest(ctx context.Context, from, to int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO friend_requests (from_user_id, to_user_id, status) VALUES ($1,$2,'pending')
		 ON CONFLICT (from_user_id, to_user_id) DO NOTHING`,
		from, to,
	)
	return err
}

func (s *FriendRepo) AcceptRequest(ctx context.Context, from, to int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx,
		`UPDATE friend_requests SET status='accepted'
		 WHERE from_user_id=$1 AND to_user_id=$2 AND status='pending'`, from, to,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	for _, pair := range [][2]int64{{from, to}, {to, from}} {
		if _, err := tx.Exec(ctx,
			`INSERT INTO friendships (user_id, friend_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			pair[0], pair[1],
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

var errNotFound = &storeError{"not found"}

type storeError struct{ msg string }

func (e *storeError) Error() string { return e.msg }

func (s *FriendRepo) AreFriends(ctx context.Context, a, b int64) (bool, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT 1 FROM friendships WHERE user_id=$1 AND friend_id=$2`, a, b,
	).Scan(&n)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *FriendRepo) ListFriends(ctx context.Context, uid int64) ([]models.User, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT u.id, u.username, u.nickname, u.avatar_url, u.created_at
		 FROM friendships f JOIN users u ON u.id = f.friend_id WHERE f.user_id=$1`, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.AvatarURL, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
