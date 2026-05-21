package code

// Gateway 10000–10099

const (
	GatewayUnauthorized  Code = 10001
	GatewayNotAuthed     Code = 10002
	GatewayInvalidFrame  Code = 10003
	GatewaySendFailed    Code = 10004
	GatewayMessageTooBig Code = 10005
	GatewayAuthTimeout   Code = 10006
)

func init() {
	register(GatewayUnauthorized, "unauthorized", "unauthorized")
	register(GatewayNotAuthed, "not_authed", "authentication required")
	register(GatewayInvalidFrame, "invalid_frame", "invalid frame")
	register(GatewaySendFailed, "send_failed", "send message failed")
	register(GatewayMessageTooBig, "message_too_large", "message too large")
	register(GatewayAuthTimeout, "auth_timeout", "auth timeout")
}
