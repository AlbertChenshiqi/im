package repo

import (
	"testing"
	"time"

	"im/pkg/models"
)

func TestSortConversationRowsPinnedFirst(t *testing.T) {
	now := time.Now()
	rows := []ConversationRow{
		{Conversation: models.Conversation{ID: "a", Pinned: false}, UpdatedAt: now},
		{Conversation: models.Conversation{ID: "b", Pinned: true}, UpdatedAt: now.Add(-time.Hour)},
		{Conversation: models.Conversation{ID: "c", Pinned: false}, UpdatedAt: now.Add(time.Hour)},
	}
	sortConversationRows(rows)
	if rows[0].ID != "b" || !rows[0].Pinned {
		t.Fatalf("pinned first: got id=%s pinned=%v", rows[0].ID, rows[0].Pinned)
	}
	if rows[1].ID != "c" {
		t.Fatalf("unpinned by time: got %s", rows[1].ID)
	}
	if rows[2].ID != "a" {
		t.Fatalf("unpinned tail: got %s", rows[2].ID)
	}
}

func TestSortConversationRowsPinnedByUpdatedAt(t *testing.T) {
	now := time.Now()
	rows := []ConversationRow{
		{Conversation: models.Conversation{ID: "old", Pinned: true}, UpdatedAt: now.Add(-2 * time.Hour)},
		{Conversation: models.Conversation{ID: "new", Pinned: true}, UpdatedAt: now},
	}
	sortConversationRows(rows)
	if rows[0].ID != "new" {
		t.Fatalf("newer pinned first: got %s", rows[0].ID)
	}
}
