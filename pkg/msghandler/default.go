package msghandler

// DefaultRegistry 内置 text / image / emoji / custom 处理器。
func DefaultRegistry() *Registry {
	return NewRegistry(
		TextHandler{},
		ImageHandler{},
		EmojiHandler{},
		CustomHandler{},
	)
}
