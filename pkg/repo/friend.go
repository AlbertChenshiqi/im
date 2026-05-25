package repo

import (
	"context"
	"database/sql"

	"im/pkg/models"
)

type FriendRepo struct {
	db *sql.DB
}

func NewFriendRepo(db *sql.DB) *FriendRepo {
	return &FriendRepo{db: db}
}

func (s *FriendRepo) CreateRequest(ctx context.Context, from, to int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT IGNORE INTO friend_requests (from_user_id, to_user_id, status) VALUES (?,?,'pending')`,
		from, to,
	)
	return err
}

func (s *FriendRepo) AcceptRequest(ctx context.Context, from, to int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`UPDATE friend_requests SET status='accepted'
		 WHERE from_user_id=? AND to_user_id=? AND status='pending'`, from, to,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errNotFound
	}
	for _, pair := range [][2]int64{{from, to}, {to, from}} {
		if _, err := tx.ExecContext(ctx,
			`INSERT IGNORE INTO friendships (user_id, friend_id) VALUES (?,?)`,
			pair[0], pair[1],
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

var errNotFound = &storeError{"not found"}

type storeError struct{ msg string }

func (e *storeError) Error() string { return e.msg }

func (s *FriendRepo) AreFriends(ctx context.Context, a, b int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM friendships WHERE user_id=? AND friend_id=?`, a, b,
	).Scan(&n)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *FriendRepo) ListFriends(ctx context.Context, uid int64) ([]models.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.username, u.nickname, u.avatar_url, u.created_at
		 FROM friendships f JOIN users u ON u.id = f.friend_id WHERE f.user_id=?`, uid,
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
