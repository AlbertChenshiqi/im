package store

import (
	"context"
	"database/sql"
	"time"

	"im/pkg/models"
	"im/pkg/msgbody"
)

func InsertMessage(ctx context.Context, db *sql.DB, m *models.Message) error {
	raw, err := msgbody.MarshalInput(m.Input)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx,
		`INSERT IGNORE INTO messages (id, conv_id, sender_id, seq, client_msg_id, input, created_at)
		 VALUES (?,?,?,?,?,?,?)`,
		m.ID, m.ConvID, m.SenderID, m.Seq, m.ClientMsgID, raw, m.CreatedAt,
	)
	return err
}

func UpdateConvMeta(ctx context.Context, db *sql.DB, convID string, seq, msgID int64, preview string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO conversation_meta (conv_id, last_seq, last_msg_id, last_preview, updated_at)
		 VALUES (?,?,?,?,CURRENT_TIMESTAMP(3))
		 ON DUPLICATE KEY UPDATE
		   last_seq=GREATEST(last_seq, VALUES(last_seq)),
		   last_msg_id=VALUES(last_msg_id),
		   last_preview=VALUES(last_preview),
		   updated_at=CURRENT_TIMESTAMP(3)`,
		convID, seq, msgID, preview,
	)
	return err
}

func MessageTime(ts int64) time.Time {
	if ts == 0 {
		return time.Now()
	}
	return time.UnixMilli(ts)
}
