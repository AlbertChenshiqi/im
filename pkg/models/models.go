package models

import "time"

const (
	ConvTypeC2C   = "c2c"   // 单聊（私信）
	ConvTypeGroup = "group" // 群聊

	// ConvTypeDirect 兼容旧库/旧事件中的 type 字段
	ConvTypeDirect = "direct"

	MsgTypeText  = "text"
	MsgTypeImage = "image"
	MsgTypeEmoji = "emoji"

	FriendPending  = "pending"
	FriendAccepted = "accepted"
	FriendRejected = "rejected"

	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type Group struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	OwnerID    int64     `json:"owner_id"`
	MaxMembers int       `json:"max_members"`
	Notice     string    `json:"notice"`
	CreatedAt  time.Time `json:"created_at"`
}

type Conversation struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	GroupID     int64  `json:"group_id,omitempty"`
	LastSeq     int64  `json:"last_seq"`
	LastPreview string `json:"last_preview"`
	Unread      int64  `json:"unread"`
	Pinned      bool   `json:"pinned"`
	Muted       bool   `json:"muted"`
}

type Message struct {
	ID          int64          `json:"id"`
	ConvID      string         `json:"conv_id"`
	SenderID    int64          `json:"sender_id"`
	Seq         int64          `json:"seq"`
	ClientMsgID string         `json:"client_msg_id,omitempty"`
	Input       []MessageInput `json:"input"`
	CreatedAt   time.Time      `json:"created_at"`
}

type BadgePayload struct {
	Type        string `json:"type"`
	ConvID      string `json:"conv_id"`
	ConvType    string `json:"conv_type"`
	Seq         int64  `json:"seq"`
	UnreadDelta int64  `json:"unread_delta"`
	Ts          int64  `json:"ts"`
}
