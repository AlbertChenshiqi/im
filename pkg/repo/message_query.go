package repo

import (
	"context"
	"database/sql"
	"time"

	"im/pkg/models"
	"im/pkg/msgbody"
)

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (s *MessageRepo) InsertMessage(ctx context.Context, m *models.Message) error {
	raw, err := msgbody.MarshalInput(m.Input)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT IGNORE INTO messages (id, conv_id, sender_id, seq, client_msg_id, input, created_at)
		 VALUES (?,?,?,?,?,?,?)`,
		m.ID, m.ConvID, m.SenderID, m.Seq, m.ClientMsgID, raw, m.CreatedAt,
	)
	return err
}

func (s *MessageRepo) ListMessages(ctx context.Context, convID string, beforeSeq int64, limit int) ([]models.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows *sql.Rows
	var err error
	if beforeSeq > 0 {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, conv_id, sender_id, seq, client_msg_id, input, created_at
			 FROM messages WHERE conv_id=? AND seq < ? ORDER BY seq DESC LIMIT ?`,
			convID, beforeSeq, limit,
		)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, conv_id, sender_id, seq, client_msg_id, input, created_at
			 FROM messages WHERE conv_id=? ORDER BY seq DESC LIMIT ?`,
			convID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Message
	for rows.Next() {
		var m models.Message
		var inputRaw string
		if err := rows.Scan(&m.ID, &m.ConvID, &m.SenderID, &m.Seq, &m.ClientMsgID, &inputRaw, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Input = msgbody.ParseInput(inputRaw)
		out = append(out, m)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}

func Now() time.Time { return time.Now() }
