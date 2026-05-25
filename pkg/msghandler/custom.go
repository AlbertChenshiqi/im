package msghandler

import (
	"encoding/json"
	"fmt"

	"im/pkg/events"
)

type customPayload struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// CustomHandler 处理 msg_type=custom；content 为 {"type":"...", "data":{...}}。
type CustomHandler struct{}

func (CustomHandler) MsgType() string { return events.MsgTypeCustom }

func (CustomHandler) Validate(content string) error {
	var p customPayload
	if err := parseJSONObject(content, &p); err != nil {
		return err
	}
	if p.Type == "" {
		return fmt.Errorf("custom.content.type required")
	}
	return nil
}

func (h CustomHandler) Normalize(content string) (string, error) {
	var p customPayload
	if err := parseJSONObject(content, &p); err != nil {
		return "", err
	}
	return marshalPayload(&p)
}
