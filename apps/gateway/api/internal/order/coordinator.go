package order

import (
	"context"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"im/apps/gateway/api/internal/protocol"
	"im/apps/message/rpc/message"
	"im/apps/message/rpc/message_client"
	"im/pkg/bizseq"
	"im/pkg/code"
	"im/pkg/redisclient"
	"im/pkg/sessionid"
)

// Coordinator 按会话分配 bizSeq 并调用 message RPC。
// 同 session 若距上一条消息 <= window，走 Redis INCR；否则用位运算直接生成 seq。
type Coordinator struct {
	messageRpc message_client.Message
	redis      *redisclient.Client
	windowMs   int64

	lanes sync.Map // convID -> *sessionLane
}

func NewCoordinator(msgRpc message_client.Message, rdb *redisclient.Client, windowMs int) *Coordinator {
	if windowMs <= 0 {
		windowMs = bizseq.SlotDivisorMs
	}
	return &Coordinator{
		messageRpc: msgRpc,
		redis:      rdb,
		windowMs:   int64(windowMs),
	}
}

type sessionLane struct {
	mu         sync.Mutex
	lastRecvMs int64
}

type pendingItem struct {
	frame        protocol.InFrame
	uid          int64
	sendTs       int64
	serverRecvMs int64
}

// Submit 分配 bizSeq 并下发消息，不再等待聚合窗口。
func (c *Coordinator) Submit(ctx context.Context, frame protocol.InFrame, uid int64) (protocol.SentOut, *protocol.ErrorOut) {
	select {
	case <-ctx.Done():
		e := protocol.NewErrorOut(code.GatewaySendFailed, "request cancelled")
		return protocol.SentOut{}, &e
	default:
	}

	sendTs := frame.SendTs
	if sendTs <= 0 {
		sendTs = time.Now().UnixMilli()
	}
	serverRecvMs := time.Now().UnixMilli()

	item := pendingItem{
		frame:        frame,
		uid:          uid,
		sendTs:       sendTs,
		serverRecvMs: serverRecvMs,
	}

	sessionID := sessionid.FromConvID(frame.ConvId)

	bizSeq, err := bizseq.Allocate(ctx, c.redis.RDB, sessionID, serverRecvMs)
	if err != nil {
		logx.Errorf("[gateway] bizseq conv=%s err=%v", frame.ConvId, err)
		e := protocol.NewErrorOut(code.GatewaySendFailed, "seq allocate failed")
		return protocol.SentOut{}, &e
	}
	
	return c.dispatchRPC(ctx, item, bizSeq)
}

func (c *Coordinator) dispatchRPC(ctx context.Context, it pendingItem, bizSeq int64) (protocol.SentOut, *protocol.ErrorOut) {
	resp, err := c.messageRpc.Send(ctx, &message.SendReq{
		SenderId:     it.uid,
		ConvId:       it.frame.ConvId,
		Content:      it.frame.Content,
		MsgType:      it.frame.MsgType,
		ClientMsgId:  it.frame.ClientMsgId,
		SendTs:       it.sendTs,
		BizSeq:       bizSeq,
		ServerRecvMs: it.serverRecvMs,
	})
	if err != nil {
		logx.Errorf("[gateway] ordered send failed uid=%d conv=%s biz_seq=%d err=%v",
			it.uid, it.frame.ConvId, bizSeq, err)
		e := protocol.NewErrorOut(code.GatewaySendFailed, err.Error())
		return protocol.SentOut{}, &e
	}
	logx.Infof("[gateway] ordered send ok uid=%d conv=%s msg_id=%d biz_seq=%d send_ts=%d",
		it.uid, it.frame.ConvId, resp.MsgId, resp.Seq, it.sendTs)
	return protocol.NewSent(resp.MsgId, resp.Seq), nil
}
