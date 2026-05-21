package convid

import "fmt"

// 会话 ID（conv_id）前缀：全局唯一，用于发消息、历史、未读等。
const (
	PrefixGroup = "group_"
	PrefixC2C   = "c2c_"
)

// Group 群聊 conv_id：group_{group_id}
func Group(groupID int64) string {
	return fmt.Sprintf("%s%d", PrefixGroup, groupID)
}

// C2C 单聊 conv_id：c2c_{较小uid}_{较大uid}（双方 uid 升序，保证唯一）
func C2C(a, b int64) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("%s%d_%d", PrefixC2C, a, b)
}

// Direct 已废弃，请用 C2C
func Direct(a, b int64) string { return C2C(a, b) }
