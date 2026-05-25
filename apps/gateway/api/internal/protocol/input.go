package protocol

import (
	"fmt"

	"im/pkg/msghandler"
)

// NormalizeInput 校验并原地规范化 frame.Input。
func NormalizeInput(f *InFrame, reg *msghandler.Registry) error {
	if len(f.Input) == 0 {
		return fmt.Errorf("input required")
	}
	for i := range f.Input {
		it := &f.Input[i]
		if it.MsgType == "" {
			return fmt.Errorf("input[%d].msgType required", i)
		}
		if it.Content == "" {
			return fmt.Errorf("input[%d].content required", i)
		}
		normalized, err := reg.Process(it.MsgType, it.Content)
		if err != nil {
			return fmt.Errorf("input[%d]: %w", i, err)
		}
		it.Content = normalized
	}
	return nil
}
