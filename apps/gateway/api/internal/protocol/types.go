package protocol

import "im/pkg/code"

// 上行帧 type 常量（JWT 在握手 query/header 校验，无需 auth 帧）
const (
	TypeSend = "send"
	TypePing = "ping"
)

// 下行帧 type 常量
const (
	TypeAuthOK        = "auth_ok"
	TypeSent          = "sent"
	TypePong          = "pong"
	TypeError = "error"
	TypeMessage       = "message"
	TypeBadge         = "badge"
	TypeNotification  = "notification"
)

// InFrame 客户端上行（先解析 type 再按分支校验）
type InFrame struct {
	Type        string `json:"type"`
	Token       string `json:"token,omitempty"`
	ConvId      string `json:"conv_id,omitempty"`
	Content     string `json:"content,omitempty"`
	MsgType     string `json:"msg_type,omitempty"`
	ClientMsgId string `json:"client_msg_id,omitempty"`
	SendTs      int64  `json:"send_ts,omitempty"` // 客户端发送时间(ms)，用于窗口内二次排序
}

type AuthOKOut struct {
	Type   string `json:"type"`
	UserID int64  `json:"user_id"`
}

type SentOut struct {
	Type  string `json:"type"`
	MsgID int64  `json:"msg_id"`
	Seq   int64  `json:"seq"`
}

type PongOut struct {
	Type string `json:"type"`
}

type ErrorOut struct {
	Type string `json:"type"`
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func NewErrorOut(c code.Code, msg ...string) ErrorOut {
	m := c.Message()
	if len(msg) > 0 && msg[0] != "" {
		m = msg[0]
	}
	return ErrorOut{Type: TypeError, Code: c.Slug(), Msg: m}
}

func NewAuthOK(userID int64) AuthOKOut {
	return AuthOKOut{Type: TypeAuthOK, UserID: userID}
}

func NewSent(msgID, seq int64) SentOut {
	return SentOut{Type: TypeSent, MsgID: msgID, Seq: seq}
}

func NewPong() PongOut {
	return PongOut{Type: TypePong}
}

