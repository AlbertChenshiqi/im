package models

// MessageInput 消息体片段（上行/事件/下行统一结构）。
type MessageInput struct {
	MsgType string `json:"msgType"`
	Content string `json:"content"`
}
