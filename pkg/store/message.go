package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"im/pkg/models"
)

func InsertMessage(ctx context.Context, pool *pgxpool.Pool, m *models.Message) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO messages (id, conv_id, sender_id, seq, client_msg_id, msg_type, content, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 ON CONFLICT (conv_id, seq) DO NOTHING`,
		m.ID, m.ConvID, m.SenderID, m.Seq, m.ClientMsgID, m.MsgType, m.Content, m.CreatedAt,
	)
	return err
}

func UpdateConvMeta(ctx context.Context, pool *pgxpool.Pool, convID string, seq, msgID int64, preview string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO conversation_meta (conv_id, last_seq, last_msg_id, last_preview, updated_at)
		 VALUES ($1,$2,$3,$4,NOW())
		 ON CONFLICT (conv_id) DO UPDATE SET
		   last_seq=GREATEST(conversation_meta.last_seq, EXCLUDED.last_seq),
		   last_msg_id=EXCLUDED.last_msg_id,
		   last_preview=EXCLUDED.last_preview,
		   updated_at=NOW()`,
		convID, seq, msgID, preview,
	)
	return err
}

func MessageTime(ts int64) time.Time {
	if ts == 0 {
		return time.Now().UTC()
	}
	return time.Unix(ts, 0).UTC()
}
