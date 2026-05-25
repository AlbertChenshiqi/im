package msghandler

import (
	"encoding/json"
	"fmt"
)

func parseJSONObject(content string, dest any) error {
	if content == "" {
		return fmt.Errorf("content required")
	}
	if err := json.Unmarshal([]byte(content), dest); err != nil {
		return fmt.Errorf("content must be valid json object: %w", err)
	}
	return nil
}

func marshalPayload(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
