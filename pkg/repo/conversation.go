package repo

import (
	"context"
	"database/sql"

	"im/pkg/convid"
	"im/pkg/models"
)

type ConversationRepo struct {
	db *sql.DB
}

func NewConversationRepo(db *sql.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

func (s *ConversationRepo) EnsureC2C(ctx context.Context, a, b int64) (string, error) {
	cid := convid.C2C(a, b)
	_, err := s.db.ExecContext(ctx,
		`INSERT IGNORE INTO conversations (id, type) VALUES (?,?)`, cid, "c2c",
	)
	if err != nil {
		return "", err
	}
	for _, uid := range []int64{a, b} {
		_, err = s.db.ExecContext(ctx,
			`INSERT IGNORE INTO conversation_members (conv_id, user_id) VALUES (?,?)`,
			cid, uid,
		)
		if err != nil {
			return "", err
		}
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT IGNORE INTO conversation_meta (conv_id) VALUES (?)`, cid,
	)
	return cid, err
}

func (s *ConversationRepo) EnsureDirect(ctx context.Context, a, b int64) (string, error) {
	return s.EnsureC2C(ctx, a, b)
}

func (s *ConversationRepo) EnsureGroupConv(ctx context.Context, groupID int64) (string, error) {
	cid := convid.Group(groupID)
	_, err := s.db.ExecContext(ctx,
		`INSERT IGNORE INTO conversations (id, type, group_id) VALUES (?,?,?)`,
		cid, "group", groupID,
	)
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT IGNORE INTO conversation_meta (conv_id) VALUES (?)`, cid,
	)
	return cid, err
}

func (s *ConversationRepo) AddGroupMembersToConv(ctx context.Context, cid string, userIDs []int64) error {
	for _, uid := range userIDs {
		if _, err := s.db.ExecContext(ctx,
			`INSERT IGNORE INTO conversation_members (conv_id, user_id) VALUES (?,?)`,
			cid, uid,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *ConversationRepo) ListForUser(ctx context.Context, uid int64) ([]models.Conversation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.type, COALESCE(c.group_id,0), COALESCE(m.last_seq,0), COALESCE(m.last_preview,''),
		        cm.pinned, cm.muted
		 FROM conversation_members cm
		 JOIN conversations c ON c.id = cm.conv_id
		 LEFT JOIN conversation_meta m ON m.conv_id = c.id
		 WHERE cm.user_id=?
		 ORDER BY (m.updated_at IS NULL), m.updated_at DESC`, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Conversation
	for rows.Next() {
		var cv models.Conversation
		if err := rows.Scan(&cv.ID, &cv.Type, &cv.GroupID, &cv.LastSeq, &cv.LastPreview, &cv.Pinned, &cv.Muted); err != nil {
			return nil, err
		}
		out = append(out, cv)
	}
	return out, rows.Err()
}

func (s *ConversationRepo) MarkRead(ctx context.Context, uid int64, convID string, seq int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE conversation_members SET last_read_seq=?, updated_at=CURRENT_TIMESTAMP(3)
		 WHERE conv_id=? AND user_id=? AND last_read_seq < ?`,
		seq, convID, uid, seq,
	)
	return err
}

func (s *ConversationRepo) UpdateMeta(ctx context.Context, convID string, seq, msgID int64, preview string) error {
	_, err := s.db.ExecContext(ctx,
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

func (s *ConversationRepo) GetConvType(ctx context.Context, convID string) (string, int64, error) {
	var typ string
	var gid sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT type, group_id FROM conversations WHERE id=?`, convID,
	).Scan(&typ, &gid)
	if err != nil {
		return "", 0, err
	}
	if gid.Valid {
		return typ, gid.Int64, nil
	}
	return typ, 0, nil
}

func (s *ConversationRepo) GetDirectMembers(ctx context.Context, convID string) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id FROM conversation_members WHERE conv_id=?`, convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
