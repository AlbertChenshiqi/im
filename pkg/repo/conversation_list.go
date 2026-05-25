package repo

import (
	"context"
	"slices"
	"time"

	"im/pkg/convid"
	"im/pkg/models"
)

type ConversationRow struct {
	models.Conversation
	PeerUserID int64
	GroupName  string
	UpdatedAt  time.Time
}

func (s *ConversationRepo) ListGroupsForUser(ctx context.Context, uid int64) ([]ConversationRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT COALESCE(c.id, CONCAT('group_', g.id)), 'group', g.id, g.name,
		        COALESCE(m.last_seq, 0), COALESCE(m.last_preview, ''),
		        COALESCE(cm.pinned, 0), COALESCE(cm.muted, gm.muted),
		        COALESCE(m.updated_at, g.created_at)
		 FROM group_members gm
		 JOIN `+"`groups`"+` g ON g.id = gm.group_id
		 LEFT JOIN conversations c ON c.type = 'group' AND c.group_id = g.id
		 LEFT JOIN conversation_meta m ON m.conv_id = COALESCE(c.id, CONCAT('group_', g.id))
		 LEFT JOIN conversation_members cm ON cm.conv_id = COALESCE(c.id, CONCAT('group_', g.id)) AND cm.user_id = ?
		 WHERE gm.user_id = ?
		 ORDER BY COALESCE(m.updated_at, g.created_at) DESC`,
		uid, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConversationRows(rows, uid, false)
}

func (s *ConversationRepo) ListDirectForUser(ctx context.Context, uid int64, recentDays int) ([]ConversationRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.type, COALESCE(c.group_id, 0), '',
		        COALESCE(m.last_seq, 0), COALESCE(m.last_preview, ''),
		        COALESCE(cm.pinned, 0), COALESCE(cm.muted, 0),
		        COALESCE(m.updated_at, cm.updated_at, c.created_at)
		 FROM conversation_members cm
		 JOIN conversations c ON c.id = cm.conv_id AND c.type IN ('c2c', 'direct')
		 LEFT JOIN conversation_meta m ON m.conv_id = c.id
		 WHERE cm.user_id = ?
		   AND (? = 0 OR COALESCE(m.updated_at, cm.updated_at, c.created_at) >= DATE_SUB(NOW(), INTERVAL ? DAY))
		 ORDER BY COALESCE(m.updated_at, cm.updated_at, c.created_at) DESC`,
		uid, recentDays, recentDays,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConversationRows(rows, uid, true)
}

func scanConversationRows(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}, self int64, isDirect bool) ([]ConversationRow, error) {
	var out []ConversationRow
	for rows.Next() {
		var r ConversationRow
		var gid int64
		var gname string
		if err := rows.Scan(
			&r.ID, &r.Type, &gid, &gname,
			&r.LastSeq, &r.LastPreview, &r.Pinned, &r.Muted, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		r.GroupID = gid
		r.GroupName = gname
		if isDirect {
			r.PeerUserID, _ = convid.C2CPeer(r.ID, self)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func MergeConversationRows(groups, directs []ConversationRow) []ConversationRow {
	out := make([]ConversationRow, 0, len(groups)+len(directs))
	out = append(out, groups...)
	out = append(out, directs...)
	sortConversationRows(out)
	return out
}

func sortConversationRows(rows []ConversationRow) {
	slices.SortFunc(rows, func(a, b ConversationRow) int {
		if a.Pinned != b.Pinned {
			if a.Pinned {
				return -1
			}
			return 1
		}
		if a.UpdatedAt.After(b.UpdatedAt) {
			return -1
		}
		if a.UpdatedAt.Before(b.UpdatedAt) {
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
}
