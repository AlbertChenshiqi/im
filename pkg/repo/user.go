package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"im/pkg/models"
	"im/pkg/sqlutil"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateUser(ctx context.Context, username, hash, nickname string) (*models.User, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, nickname) VALUES (?,?,?)`,
		username, hash, nickname,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (id int64, hash string, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT id, password_hash FROM users WHERE username=?`, username,
	).Scan(&id, &hash)
	return
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, avatar_url, created_at FROM users WHERE id=?`, id,
	).Scan(&u.ID, &u.Username, &u.Nickname, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) EnsureDevUser(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	u, err := r.GetByID(ctx, id)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	username := fmt.Sprintf("dev_%d", id)
	nickname := fmt.Sprintf("Dev %d", id)
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, password_hash, nickname) VALUES (?,?,'',?)
		 ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP(3)`,
		id, username, nickname,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *UserRepo) FindMissingIDs(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := `SELECT id FROM users WHERE id IN (` + sqlutil.Placeholders(len(ids)) + `)`
	rows, err := r.db.QueryContext(ctx, q, sqlutil.Int64Args(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	exists := make(map[int64]bool, len(ids))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		exists[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var missing []int64
	for _, id := range ids {
		if id > 0 && !exists[id] {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

func (r *UserRepo) BatchGetByIDs(ctx context.Context, ids []int64) ([]*models.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	q := `SELECT id, username, nickname, avatar_url, created_at FROM users WHERE id IN (` +
		sqlutil.Placeholders(len(ids)) + `)`
	rows, err := r.db.QueryContext(ctx, q, sqlutil.Int64Args(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.AvatarURL, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, rows.Err()
}

func (r *UserRepo) UpsertDevice(ctx context.Context, userID int64, deviceID, pushToken, platform string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_devices (user_id, device_id, push_token, platform, updated_at)
		 VALUES (?,?,?,?,CURRENT_TIMESTAMP(3))
		 ON DUPLICATE KEY UPDATE push_token=VALUES(push_token), platform=VALUES(platform), updated_at=CURRENT_TIMESTAMP(3)`,
		userID, deviceID, pushToken, platform,
	)
	return err
}
