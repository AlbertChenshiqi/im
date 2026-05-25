package msgbody

import (
	"encoding/json"
	"strings"

	"im/pkg/events"
	"im/pkg/models"
)

type inputPayload struct {
	Input []models.MessageInput `json:"input"`
}

// MarshalInput 序列化落库的 input JSON。
func MarshalInput(input []models.MessageInput) (string, error) {
	if len(input) == 0 {
		return "", errEmpty
	}
	b, err := json.Marshal(inputPayload{Input: input})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParseInput 从库表 input 列还原片段列表。
func ParseInput(raw string) []models.MessageInput {
	if raw == "" {
		return nil
	}
	var wrapped inputPayload
	if err := json.Unmarshal([]byte(raw), &wrapped); err == nil && len(wrapped.Input) > 0 {
		return wrapped.Input
	}
	return nil
}

// ChatTag 按 input 片段选择 im_chat Tag。
func ChatTag(convType string, input []models.MessageInput) string {
	for _, it := range input {
		if it.MsgType == events.MsgTypeRecall {
			return events.TagChatRecall
		}
		if it.MsgType == events.MsgTypeCustom || strings.HasPrefix(it.MsgType, "custom_") {
			return events.TagChatCustom
		}
	}
	if convType == models.ConvTypeGroup {
		return events.TagChatGroup
	}
	return events.TagChatC2C
}

// Preview 会话列表摘要。
func Preview(input []models.MessageInput) string {
	if len(input) == 0 {
		return ""
	}
	first := input[0]
	var p map[string]any
	_ = json.Unmarshal([]byte(first.Content), &p)
	switch first.MsgType {
	case models.MsgTypeText:
		if s, ok := p["text"].(string); ok {
			return s
		}
	case models.MsgTypeImage:
		return "[图片]"
	case models.MsgTypeEmoji:
		if s, ok := p["emoji"].(string); ok {
			return s
		}
	}
	if len(input) > 1 {
		return "[图文消息]"
	}
	return first.Content
}

var errEmpty = errString("input required")

type errString string

func (e errString) Error() string { return string(e) }
