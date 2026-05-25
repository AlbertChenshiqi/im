package msghandler

import (
	"fmt"
	"strings"

	"im/pkg/models"
)

type textPayload struct {
	Text string `json:"text"`
}

type TextHandler struct{}

func (TextHandler) MsgType() string { return models.MsgTypeText }

func (TextHandler) Validate(content string) error {
	var p textPayload
	if err := parseJSONObject(content, &p); err != nil {
		return err
	}
	if strings.TrimSpace(p.Text) == "" {
		return fmt.Errorf("text.content.text required")
	}
	return nil
}

func (h TextHandler) Normalize(content string) (string, error) {
	var p textPayload
	if err := parseJSONObject(content, &p); err != nil {
		return "", err
	}
	p.Text = strings.TrimSpace(p.Text)
	return marshalPayload(&p)
}
