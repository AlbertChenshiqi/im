package msgcore

import (
	"context"
	"fmt"
	"time"

	"im/apps/group/rpc/group"
	"im/apps/group/rpc/group_client"
	"im/pkg/convid"
	"im/pkg/events"
	"im/pkg/models"
	"im/pkg/redisclient"
	"im/pkg/repo"
	"im/pkg/rocketmq"
	"im/pkg/sessionid"
	"im/pkg/snowflake"
)

type Sender struct {
	RDB      *redisclient.Client
	Producer *rocketmq.Producer
	SF       *snowflake.Generator
	GroupRpc group_client.Group
	ConvRepo *repo.ConversationRepo
}

type SendInput struct {
	SenderID     int64
	ConvID       string
	Content      string
	MsgType      string
	ClientMsgID  string
	BizSeq       int64
	SendTs       int64
	ServerRecvMs int64
}

type SendResult struct {
	MsgID int64
	Seq   int64
}

func (s *Sender) Send(ctx context.Context, in SendInput) (*SendResult, error) {
	if in.MsgType == "" {
		in.MsgType = models.MsgTypeText
	}
	if in.BizSeq <= 0 {
		return nil, fmt.Errorf("biz_seq required")
	}
	if dup, err := s.RDB.CheckDedupe(ctx, in.ClientMsgID, 0); err == nil && dup {
		return nil, fmt.Errorf("duplicate client_msg_id")
	}
	convType, groupID, err := resolveConv(in.ConvID)
	if err != nil {
		return nil, err
	}
	var recipients []int64
	if convType == models.ConvTypeC2C {
		recipients, err = c2cMembers(in.ConvID, in.SenderID)
		if err != nil {
			return nil, err
		}
		if s.ConvRepo != nil && len(recipients) == 2 {
			_, _ = s.ConvRepo.EnsureC2C(ctx, recipients[0], recipients[1])
		}
	} else if s.GroupRpc != nil {
		resp, err := s.GroupRpc.IsMember(ctx, &group.IsMemberReq{GroupId: groupID, UserId: in.SenderID})
		if err != nil || resp == nil || !resp.Ok {
			return nil, fmt.Errorf("not a group member")
		}
	}
	msgID := s.SF.Next()
	if in.ClientMsgID != "" {
		_, _ = s.RDB.CheckDedupe(ctx, in.ClientMsgID, msgID)
	}
	sid := sessionid.FromConvID(in.ConvID)
	ts := time.Now().Unix()
	if in.SendTs > 0 {
		ts = in.SendTs / 1000
	}
	evt := events.MessageSendEvent{
		MsgID: msgID, ConvID: in.ConvID, SessionID: sid, ConvType: convType, GroupID: groupID,
		SenderID: in.SenderID, BizSeq: in.BizSeq, Seq: in.BizSeq,
		SendTs: in.SendTs, ServerRecvMs: in.ServerRecvMs,
		ClientMsgID: in.ClientMsgID, MsgType: in.MsgType, Content: in.Content, Ts: ts,
		RecipientIDs: recipients,
	}
	tag := events.ChatTagForSend(evt)
	// 同 sessionId 作 Message Key，保证同会话 RocketMQ 分区内有序
	if err := s.Producer.PublishJSON(ctx, events.TopicChat, tag, sid, evt); err != nil {
		return nil, err
	}
	if err := s.Producer.PublishJSON(ctx, events.TopicChatPersist, events.TagChatPersistStore, sid, evt); err != nil {
		return nil, err
	}
	return &SendResult{MsgID: msgID, Seq: in.BizSeq}, nil
}

func resolveConv(convID string) (string, int64, error) {
	if gid, ok := convid.ParseGroupID(convID); ok {
		return models.ConvTypeGroup, gid, nil
	}
	if convid.IsC2C(convID) {
		return models.ConvTypeC2C, 0, nil
	}
	return "", 0, fmt.Errorf("unknown conversation id")
}

func c2cMembers(convID string, self int64) ([]int64, error) {
	a, b, ok := convid.ParseC2C(convID)
	if !ok {
		return nil, fmt.Errorf("bad c2c conv id")
	}
	if self != a && self != b {
		return nil, fmt.Errorf("not member")
	}
	return []int64{a, b}, nil
}
