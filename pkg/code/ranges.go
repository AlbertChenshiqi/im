package code

// 错误码分段：与 apps API 端口 10XYZ 对齐，每域预留 100 个码 (XYZ00–XYZ99)。
//
//	10000–10099  gateway
//	10100–10199  user
//	10200–10299  friend
//	10300–10399  group
//	10400–10499  conversation
//	10500–10599  message
//	10600–10699  notification
//	10700–10799  push
//	10800–10899  cron
//	1–999        common（跨服务通用）
const (
	RangeCommonMin = 1
	RangeCommonMax = 999

	RangeGatewayMin = 10000
	RangeGatewayMax = 10099

	RangeUserMin = 10100
	RangeUserMax = 10199

	RangeFriendMin = 10200
	RangeFriendMax = 10299

	RangeGroupMin = 10300
	RangeGroupMax = 10399

	RangeConversationMin = 10400
	RangeConversationMax = 10499

	RangeMessageMin = 10500
	RangeMessageMax = 10599

	RangeNotificationMin = 10600
	RangeNotificationMax = 10699

	RangePushMin = 10700
	RangePushMax = 10799

	RangeCronMin = 10800
	RangeCronMax = 10899
)

// Segment 返回码所在分段名称
func Segment(c Code) string {
	n := int(c)
	switch {
	case n >= RangeGatewayMin && n <= RangeGatewayMax:
		return "gateway"
	case n >= RangeUserMin && n <= RangeUserMax:
		return "user"
	case n >= RangeFriendMin && n <= RangeFriendMax:
		return "friend"
	case n >= RangeGroupMin && n <= RangeGroupMax:
		return "group"
	case n >= RangeConversationMin && n <= RangeConversationMax:
		return "conversation"
	case n >= RangeMessageMin && n <= RangeMessageMax:
		return "message"
	case n >= RangeNotificationMin && n <= RangeNotificationMax:
		return "notification"
	case n >= RangePushMin && n <= RangePushMax:
		return "push"
	case n >= RangeCronMin && n <= RangeCronMax:
		return "cron"
	case n >= RangeCommonMin && n <= RangeCommonMax:
		return "common"
	default:
		return "unknown"
	}
}

// InRange 校验码是否落在指定分段内
func InRange(c Code, min, max int) bool {
	n := int(c)
	return n >= min && n <= max
}
