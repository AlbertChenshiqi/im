package protocol

import "fmt"

func (f *InFrame) Validate() error {
	if f.Type == "" {
		return fmt.Errorf("missing type")
	}
	switch f.Type {
	case TypeSend:
		if f.ConvId == "" {
			return fmt.Errorf("conv_id required")
		}
		if len(f.Input) == 0 {
			return fmt.Errorf("input required")
		}
	case TypePing:
		// no fields
	default:
		return fmt.Errorf("unknown type: %s", f.Type)
	}
	return nil
}
