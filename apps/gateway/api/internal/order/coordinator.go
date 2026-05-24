package order

import (
	"context"
	"sort"
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

// Coordinator 按会话聚合短时窗口，按 sendTs 规整后顺序分配 bizSeq 并调用 message RPC。
type Coordinator struct {
	messageRpc message_client.Message
	redis      *redisclient.Client
	window     time.Duration

	mu      sync.Mutex
	buffers map[string]*convBuffer
}

func NewCoordinator(msgRpc message_client.Message, rdb *redisclient.Client, windowMs int) *Coordinator {
	if windowMs <= 0 {
		windowMs = 200
	}
	return &Coordinator{
		messageRpc: msgRpc,
		redis:      rdb,
		window:     time.Duration(windowMs) * time.Millisecond,
		buffers:    make(map[string]*convBuffer),
	}
}

type sendReply struct {
	out protocol.SentOut
	err *protocol.ErrorOut
}

type pendingItem struct {
	frame        protocol.InFrame
	uid          int64
	sendTs       int64
	serverRecvMs int64
	reply        chan sendReply
}

// Submit 将发送请求放入会话缓冲，窗口结束后批量规整并下发。
func (c *Coordinator) Submit(ctx context.Context, frame protocol.InFrame, uid int64) (protocol.SentOut, *protocol.ErrorOut) {
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
		reply:        make(chan sendReply, 1),
	}

	c.mu.Lock()
	buf := c.buffers[frame.ConvId]
	if buf == nil {
		buf = &convBuffer{coord: c, convID: frame.ConvId}
		c.buffers[frame.ConvId] = buf
	}
	buf.add(item)
	c.mu.Unlock()

	select {
	case <-ctx.Done():
		e := protocol.NewErrorOut(code.GatewaySendFailed, "request cancelled")
		return protocol.SentOut{}, &e
	case rep := <-item.reply:
		return rep.out, rep.err
	}
}

func (c *Coordinator) flush(convID string) {
	c.mu.Lock()
	buf := c.buffers[convID]
	delete(c.buffers, convID)
	c.mu.Unlock()
	if buf == nil {
		return
	}

	items := buf.take()
	if len(items) == 0 {
		return
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].sendTs != items[j].sendTs {
			return items[i].sendTs < items[j].sendTs
		}
		return items[i].serverRecvMs < items[j].serverRecvMs
	})

	sessionID := sessionid.FromConvID(convID)
	ctx := context.Background()

	for _, it := range items {
		bizSeq, err := bizseq.Allocate(ctx, c.redis.RDB, sessionID, it.serverRecvMs)
		if err != nil {
			logx.Errorf("[gateway] bizseq conv=%s err=%v", convID, err)
			e := protocol.NewErrorOut(code.GatewaySendFailed, "seq allocate failed")
			it.reply <- sendReply{err: &e}
			continue
		}

		out, errOut := c.dispatchRPC(ctx, it, bizSeq)
		it.reply <- sendReply{out: out, err: errOut}
	}
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

type convBuffer struct {
	coord  *Coordinator
	convID string
	mu     sync.Mutex
	items  []pendingItem
	timer  *time.Timer
}

func (b *convBuffer) add(it pendingItem) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = append(b.items, it)
	if b.timer == nil {
		b.timer = time.AfterFunc(b.coord.window, func() {
			b.coord.flush(b.convID)
		})
	}
}

func (b *convBuffer) take() []pendingItem {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	out := b.items
	b.items = nil
	return out
}
