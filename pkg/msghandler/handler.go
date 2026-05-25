package msghandler

import "fmt"

// Handler 按 msg_type 处理上行消息体：校验 content 并规范化为落库/下发的 JSON 字符串。
// 新增消息类型时实现本接口并 Register 到 Registry。
type Handler interface {
	MsgType() string
	Validate(content string) error
	Normalize(content string) (string, error)
}

// Registry 按 msg_type 路由到对应 Handler。
type Registry struct {
	byType map[string]Handler
}

func NewRegistry(handlers ...Handler) *Registry {
	r := &Registry{byType: make(map[string]Handler, len(handlers))}
	for _, h := range handlers {
		r.Register(h)
	}
	return r
}

func (r *Registry) Register(h Handler) {
	if h == nil {
		return
	}
	r.byType[h.MsgType()] = h
}

func (r *Registry) Get(msgType string) (Handler, bool) {
	h, ok := r.byType[msgType]
	return h, ok
}

// Process 校验并规范化单条 input。
func (r *Registry) Process(msgType, content string) (string, error) {
	if msgType == "" {
		return "", fmt.Errorf("msg_type required")
	}
	h, ok := r.Get(msgType)
	if !ok {
		return "", fmt.Errorf("unsupported msg_type: %s", msgType)
	}
	if err := h.Validate(content); err != nil {
		return "", err
	}
	return h.Normalize(content)
}
