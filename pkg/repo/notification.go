package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Category  string    `json:"category"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func (s *NotificationRepo) Create(ctx context.Context, userID int64, title, body, category string) (*Notification, error) {
	var n Notification
	err := s.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, title, body, category)
		 VALUES ($1,$2,$3,$4) RETURNING id, user_id, title, body, category, read, created_at`,
		userID, title, body, category,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Category, &n.Read, &n.CreatedAt)
	return &n, err
}

func (s *NotificationRepo) List(ctx context.Context, userID int64, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, title, body, category, read, created_at
		 FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Category, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *NotificationRepo) MarkRead(ctx context.Context, userID, id int64) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE notifications SET read=TRUE WHERE id=$1 AND user_id=$2`, id, userID,
	)
	return err
}
