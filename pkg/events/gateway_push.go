package events

import (
	"context"
	"encoding/json"
	"strconv"

	"im/pkg/models"
	"im/pkg/rocketmq"
)

type WSMessagePush struct {
	Type        string `json:"type"` // message
	MsgID       int64  `json:"msg_id"`
	ConvID      string `json:"conv_id"`
	ConvType    string `json:"conv_type"`
	SenderID    int64  `json:"sender_id"`
	Seq         int64  `json:"seq"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
	MsgType     string `json:"msg_type"`
	Content     string `json:"content"`
	Ts          int64  `json:"ts"`
}

type WSBadgePush struct {
	Type        string `json:"type"` // badge
	ConvID      string `json:"conv_id"`
	ConvType    string `json:"conv_type"`
	Seq         int64  `json:"seq"`
	UnreadDelta int64  `json:"unread_delta"`
	UnreadTotal int64  `json:"unread_total"`
	Ts          int64  `json:"ts"`
}

type WSNotificationPush struct {
	Type     string `json:"type"` // notification
	Title    string `json:"title"`
	Body     string `json:"body"`
	Category string `json:"category"`
	Ts       int64  `json:"ts"`
}

func MessageFrame(evt MessageSendEvent) WSMessagePush {
	msgType := evt.MsgType
	if msgType == "" {
		msgType = models.MsgTypeText
	}
	return WSMessagePush{
		Type: "message", MsgID: evt.MsgID, ConvID: evt.ConvID, ConvType: evt.ConvType,
		SenderID: evt.SenderID, Seq: evt.Seq, ClientMsgID: evt.ClientMsgID,
		MsgType: msgType, Content: evt.Content, Ts: evt.Ts,
	}
}

func BadgeFrame(e InboxUpdatedEvent, unreadTotal int64) WSBadgePush {
	return WSBadgePush{
		Type: "badge", ConvID: e.ConvID, ConvType: e.ConvType, Seq: e.Seq,
		UnreadDelta: e.UnreadDelta, UnreadTotal: unreadTotal, Ts: e.Ts,
	}
}

func NotificationFrame(e NotificationEvent, ts int64) WSNotificationPush {
	return WSNotificationPush{
		Type: "notification", Title: e.Title, Body: e.Body, Category: e.Category, Ts: ts,
	}
}

// PublishGatewayPush 写入 im / gateway_push；value 含 user_id，gateway 按 JSON type 路由 WS 帧。
func PublishGatewayPush(ctx context.Context, p *rocketmq.Producer, uid int64, payload any) error {
	pb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(pb, &m); err != nil {
		return err
	}
	m["user_id"] = uid
	return p.PublishJSON(ctx, TopicSync, TagSyncGateway, strconv.FormatInt(uid, 10), m)
}
