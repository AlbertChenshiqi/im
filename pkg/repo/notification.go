package repo

import (
	"context"
	"database/sql"
	"time"
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
	db *sql.DB
}

func NewNotificationRepo(db *sql.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (s *NotificationRepo) Create(ctx context.Context, userID int64, title, body, category string) (*Notification, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, title, body, category) VALUES (?,?,?,?)`,
		userID, title, body, category,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	var n Notification
	err = s.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, body, category, is_read, created_at FROM notifications WHERE id=?`, id,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Category, &n.Read, &n.CreatedAt)
	return &n, err
}

func (s *NotificationRepo) List(ctx context.Context, userID int64, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, title, body, category, is_read, created_at
		 FROM notifications WHERE user_id=? ORDER BY created_at DESC LIMIT ?`, userID, limit,
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
	_, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET is_read=1 WHERE id=? AND user_id=?`, id, userID,
	)
	return err
}
