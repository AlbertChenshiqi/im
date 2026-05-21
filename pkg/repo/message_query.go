package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/models"
)

type MessageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

func (s *MessageRepo) InsertMessage(ctx context.Context, m *models.Message) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO messages (id, conv_id, sender_id, seq, client_msg_id, msg_type, content, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 ON CONFLICT (conv_id, seq) DO NOTHING`,
		m.ID, m.ConvID, m.SenderID, m.Seq, m.ClientMsgID, m.MsgType, m.Content, m.CreatedAt,
	)
	return err
}

func (s *MessageRepo) ListMessages(ctx context.Context, convID string, beforeSeq int64, limit int) ([]models.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := `SELECT id, conv_id, sender_id, seq, client_msg_id, msg_type, content, created_at
	      FROM messages WHERE conv_id=$1`
	args := []any{convID}
	if beforeSeq > 0 {
		q += ` AND seq < $2 ORDER BY seq DESC LIMIT $3`
		args = append(args, beforeSeq, limit)
	} else {
		q += ` ORDER BY seq DESC LIMIT $2`
		args = append(args, limit)
	}
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ConvID, &m.SenderID, &m.Seq, &m.ClientMsgID, &m.MsgType, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	// reverse to ascending
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}

func Now() time.Time { return time.Now().UTC() }
