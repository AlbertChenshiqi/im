package events

import "im/pkg/models"

// RocketMQ Topic：按主业务域拆分。
const (
	TopicChat        = "im_chat"         // 聊天主消息（扇出：未读、实时推送等）
	TopicChatPersist = "im_chat_persist" // 异步落库（按 sessionId 保序）
	TopicPush        = "im_push"         // 离线推送（离线消息、系统公告）
	TopicSync        = "im_sync"         // 状态同步（已读、上下线、好友变更、实时下行）
)

// im_chat Tag
const (
	TagChatC2C       = "c2c"
	TagChatGroup     = "group"
	TagChatRecall    = "recall"
	TagChatCustom    = "custom"
	TagChatPersisted     = "persisted"
	TagChatPersistStore  = "store" // im_chat_persist 落库任务
)

// im_push Tag
const (
	TagPushOffline      = "offline_message"
	TagSystemAnnounce   = "system_announce"
	TagNotificationSystem = TagSystemAnnounce // 兼容旧名
)

// im_sync Tag
const (
	TagSyncRead         = "read"
	TagSyncGateway      = "gateway_push"
	TagSyncOnline       = "online"
	TagSyncFriend       = "friend"
	TagSyncGroupMember  = "group_member"
	TagInboxUpdated     = TagSyncRead
	TagGatewayPush      = TagSyncGateway
	TagGroupMember      = TagSyncGroupMember
)

// ChatSubscribeAll 聊天域消费方订阅表达式（落库、未读、实时推送等）。
const ChatSubscribeAll = "c2c || group || custom || recall"

const MsgTypeRecall = "recall"
const MsgTypeCustom = "custom"

type MessageSendEvent struct {
	MsgID        int64  `json:"msg_id"`
	ConvID       string `json:"conv_id"`
	SessionID    string `json:"session_id"`
	ConvType     string `json:"conv_type"`
	GroupID      int64  `json:"group_id,omitempty"`
	SenderID     int64  `json:"sender_id"`
	BizSeq       int64  `json:"biz_seq"`
	Seq          int64  `json:"seq"` // 与 biz_seq 相同，兼容消费方
	SendTs       int64  `json:"send_ts"`
	ServerRecvMs int64  `json:"server_recv_ms"`
	ClientMsgID  string                `json:"client_msg_id"`
	Input        []models.MessageInput `json:"input"`
	Ts           int64                 `json:"ts"`
	RecipientIDs []int64 `json:"recipient_ids,omitempty"`
}

type MessageRecallEvent struct {
	MsgID    int64  `json:"msg_id"`
	ConvID   string `json:"conv_id"`
	ConvType string `json:"conv_type"`
	GroupID  int64  `json:"group_id,omitempty"`
	SenderID int64  `json:"sender_id"`
	Seq      int64  `json:"seq"`
	Ts       int64  `json:"ts"`
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
	UserID int64  `json:"user_id"`
	ConvID string `json:"conv_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Count  int    `json:"count"`
	Ts     int64  `json:"ts"`
}

type OnlineStatusEvent struct {
	UserID int64  `json:"user_id"`
	Online bool   `json:"online"`
	Ts     int64  `json:"ts"`
	GatewayInstance string `json:"gateway_instance,omitempty"`
}

type FriendChangeEvent struct {
	UserID   int64  `json:"user_id"`
	PeerID   int64  `json:"peer_id"`
	Action   string `json:"action"` // add | remove | accept
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
