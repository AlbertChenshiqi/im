package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/models"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) CreateUser(ctx context.Context, username, hash, nickname string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, nickname) VALUES ($1,$2,$3)
		 RETURNING id, username, nickname, avatar_url, created_at`,
		username, hash, nickname,
	).Scan(&u.ID, &u.Username, &u.Nickname, &u.AvatarURL, &u.CreatedAt)
	return &u, err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (id int64, hash string, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE username=$1`, username,
	).Scan(&id, &hash)
	return
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, nickname, avatar_url, created_at FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Username, &u.Nickname, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// EnsureDevUser 开发用：保证 user_id 存在，不存在则创建占位用户
func (r *UserRepo) EnsureDevUser(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	u, err := r.GetByID(ctx, id)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	username := fmt.Sprintf("dev_%d", id)
	nickname := fmt.Sprintf("Dev %d", id)
	var created models.User
	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (id, username, password_hash, nickname)
		 VALUES ($1, $2, '', $3)
		 ON CONFLICT (id) DO UPDATE SET updated_at = NOW()
		 RETURNING id, username, nickname, avatar_url, created_at`,
		id, username, nickname,
	).Scan(&created.ID, &created.Username, &created.Nickname, &created.AvatarURL, &created.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// FindMissingIDs 返回 users 表中不存在的 id
func (r *UserRepo) FindMissingIDs(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT id FROM users WHERE id = ANY($1)`, ids)
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
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, nickname, avatar_url, created_at FROM users WHERE id = ANY($1)`, ids,
	)
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
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_devices (user_id, device_id, push_token, platform, updated_at)
		 VALUES ($1,$2,$3,$4,NOW())
		 ON CONFLICT (user_id, device_id) DO UPDATE SET push_token=$3, platform=$4, updated_at=NOW()`,
		userID, deviceID, pushToken, platform,
	)
	return err
}
