package msghandler

import (
	"fmt"
	"strings"

	"im/pkg/models"
)

type imagePayload struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type ImageHandler struct{}

func (ImageHandler) MsgType() string { return models.MsgTypeImage }

func (ImageHandler) Validate(content string) error {
	var p imagePayload
	if err := parseJSONObject(content, &p); err != nil {
		return err
	}
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("image.content.url required")
	}
	return nil
}

func (h ImageHandler) Normalize(content string) (string, error) {
	var p imagePayload
	if err := parseJSONObject(content, &p); err != nil {
		return "", err
	}
	p.URL = strings.TrimSpace(p.URL)
	return marshalPayload(&p)
}
