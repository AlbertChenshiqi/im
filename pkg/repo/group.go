package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"im/pkg/convid"
	"im/pkg/models"
)

var ErrTooManyMembers = &memberError{"too many members"}

type memberError struct{ msg string }

func (e *memberError) Error() string { return e.msg }

type ErrMembersNotFound struct {
	IDs []int64
}

func (e *ErrMembersNotFound) Error() string {
	parts := make([]string, len(e.IDs))
	for i, id := range e.IDs {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return fmt.Sprintf("users not found: %s", strings.Join(parts, ","))
}

type GroupRepo struct {
	db *sql.DB
}

func NewGroupRepo(db *sql.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

func (s *GroupRepo) CreateGroup(ctx context.Context, name string, ownerID int64, memberIDs []int64) (*models.Group, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx,
		`INSERT INTO `+"`groups`"+` (name, owner_id) VALUES (?,?)`, name, ownerID,
	)
	if err != nil {
		return nil, err
	}
	gid, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	var g models.Group
	err = tx.QueryRowContext(ctx,
		`SELECT id, name, owner_id, max_members, notice, created_at FROM `+"`groups`"+` WHERE id=?`, gid,
	).Scan(&g.ID, &g.Name, &g.OwnerID, &g.MaxMembers, &g.Notice, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	all := []int64{ownerID}
	seen := map[int64]bool{ownerID: true}
	for _, uid := range memberIDs {
		if uid > 0 && !seen[uid] {
			seen[uid] = true
			all = append(all, uid)
		}
	}
	if len(all) > g.MaxMembers {
		return nil, ErrTooManyMembers
	}
	for _, uid := range all {
		role := models.RoleMember
		if uid == ownerID {
			role = models.RoleOwner
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT IGNORE INTO group_members (group_id, user_id, role) VALUES (?,?,?)`,
			g.ID, uid, role,
		); err != nil {
			return nil, err
		}
	}
	cid := convid.Group(g.ID)
	if _, err := tx.ExecContext(ctx,
		`INSERT IGNORE INTO conversations (id, type, group_id) VALUES (?,?,?)`,
		cid, "group", g.ID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT IGNORE INTO conversation_meta (conv_id) VALUES (?)`, cid,
	); err != nil {
		return nil, err
	}
	for _, uid := range all {
		if _, err := tx.ExecContext(ctx,
			`INSERT IGNORE INTO conversation_members (conv_id, user_id) VALUES (?,?)`,
			cid, uid,
		); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *GroupRepo) GetGroup(ctx context.Context, id int64) (*models.Group, error) {
	var g models.Group
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, owner_id, max_members, notice, created_at FROM `+"`groups`"+` WHERE id=?`, id,
	).Scan(&g.ID, &g.Name, &g.OwnerID, &g.MaxMembers, &g.Notice, &g.CreatedAt)
	return &g, err
}

func (s *GroupRepo) IsMember(ctx context.Context, groupID, userID int64) (ok bool, muted bool, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT 1, muted FROM group_members WHERE group_id=? AND user_id=?`, groupID, userID,
	).Scan(&ok, &muted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	return
}

func (s *GroupRepo) MemberCount(ctx context.Context, groupID int64) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM group_members WHERE group_id=?`, groupID).Scan(&n)
	return n, err
}

func (s *GroupRepo) AddMembers(ctx context.Context, groupID int64, userIDs []int64) error {
	g, err := s.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}
	cnt, _ := s.MemberCount(ctx, groupID)
	if cnt+len(userIDs) > g.MaxMembers {
		return ErrTooManyMembers
	}
	cid := convid.Group(groupID)
	for _, uid := range userIDs {
		if _, err := s.db.ExecContext(ctx,
			`INSERT IGNORE INTO group_members (group_id, user_id, role) VALUES (?,?,'member')`,
			groupID, uid,
		); err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx,
			`INSERT IGNORE INTO conversation_members (conv_id, user_id) VALUES (?,?)`,
			cid, uid,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *GroupRepo) ListMembers(ctx context.Context, groupID, cursor int64, limit int) ([]int64, int64, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_id FROM group_members
		 WHERE group_id=? AND user_id > ? ORDER BY user_id LIMIT ?`,
		groupID, cursor, limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var ids []int64
	var last int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
		last = id
	}
	return ids, last, rows.Err()
}
