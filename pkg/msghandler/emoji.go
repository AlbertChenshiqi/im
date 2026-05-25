package msghandler

import (
	"fmt"
	"strings"

	"im/pkg/models"
)

type emojiPayload struct {
	Emoji string `json:"emoji"`
}

type EmojiHandler struct{}

func (EmojiHandler) MsgType() string { return models.MsgTypeEmoji }

func (EmojiHandler) Validate(content string) error {
	var p emojiPayload
	if err := parseJSONObject(content, &p); err != nil {
		return err
	}
	if strings.TrimSpace(p.Emoji) == "" {
		return fmt.Errorf("emoji.content.emoji required")
	}
	return nil
}

func (h EmojiHandler) Normalize(content string) (string, error) {
	var p emojiPayload
	if err := parseJSONObject(content, &p); err != nil {
		return "", err
	}
	p.Emoji = strings.TrimSpace(p.Emoji)
	return marshalPayload(&p)
}
