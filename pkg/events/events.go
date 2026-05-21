package events

import "im/pkg/models"

const (
	TopicMessageSend       = "im.message.send"
	TopicMessagePersisted  = "im.message.persisted"
	TopicInboxUpdated      = "im.inbox.updated"
	TopicGroupMember       = "im.group.member"
	TopicNotificationSystem = "im.notification.system"
	TopicPushOffline       = "im.push.offline"
)

type MessageSendEvent struct {
	MsgID       int64  `json:"msg_id"`
	ConvID      string `json:"conv_id"`
	ConvType    string `json:"conv_type"`
	GroupID     int64  `json:"group_id,omitempty"`
	SenderID    int64  `json:"sender_id"`
	Seq         int64  `json:"seq"`
	ClientMsgID string `json:"client_msg_id"`
	MsgType     string `json:"msg_type"`
	Content     string `json:"content"`
	Ts          int64  `json:"ts"`
	// RecipientIDs 单聊双方；群聊由 worker 按 group_members + IsOnline 扇出
	RecipientIDs []int64 `json:"recipient_ids,omitempty"`
}

type InboxUpdatedEvent struct {
	UserID      int64  `json:"user_id"`
	ConvID      string `json:"conv_id"`
	ConvType    string `json:"conv_type"`
	Seq         int64  `json:"seq"`
	UnreadDelta int64  `json:"unread_delta"`
	Ts          int64  `json:"ts"`
}

type PushOfflineEvent struct {
	UserID   int64  `json:"user_id"`
	ConvID   string `json:"conv_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Count    int    `json:"count"`
	Ts       int64  `json:"ts"`
}

type GroupMemberEvent struct {
	GroupID int64  `json:"group_id"`
	UserID  int64  `json:"user_id"`
	Action  string `json:"action"`
}

type NotificationEvent struct {
	UserID   int64  `json:"user_id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Category string `json:"category"`
}

func BadgeFromInbox(e InboxUpdatedEvent) models.BadgePayload {
	return models.BadgePayload{
		Type:        "badge",
		ConvID:      e.ConvID,
		ConvType:    e.ConvType,
		Seq:         e.Seq,
		UnreadDelta: e.UnreadDelta,
		Ts:          e.Ts,
	}
}
