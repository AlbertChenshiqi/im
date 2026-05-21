package convid

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseGroupID 从 group_{id}（兼容旧 g_{id}）解析群 ID。
func ParseGroupID(convID string) (groupID int64, ok bool) {
	for _, p := range []string{PrefixGroup, "g_"} {
		if strings.HasPrefix(convID, p) {
			_, err := fmt.Sscanf(convID, p+"%d", &groupID)
			if err == nil && groupID > 0 {
				return groupID, true
			}
		}
	}
	return 0, false
}

// ParseC2C 解析 c2c_{a}_{b}（兼容旧 d_{a}_{b}），返回排序后的两端 uid。
func ParseC2C(convID string) (a, b int64, ok bool) {
	rest, ok := trimC2CPrefix(convID)
	if !ok {
		return 0, 0, false
	}
	parts := strings.Split(rest, "_")
	if len(parts) != 2 {
		return 0, 0, false
	}
	a, err1 := strconv.ParseInt(parts[0], 10, 64)
	b, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil || a <= 0 || b <= 0 {
		return 0, 0, false
	}
	if a > b {
		a, b = b, a
	}
	return a, b, true
}

// C2CPeer 从 c2c 会话解析对端 uid。
func C2CPeer(convID string, self int64) (peer int64, ok bool) {
	a, b, ok := ParseC2C(convID)
	if !ok {
		return 0, false
	}
	switch self {
	case a:
		return b, true
	case b:
		return a, true
	default:
		return 0, false
	}
}

// DirectPeer 已废弃，请用 C2CPeer
func DirectPeer(convID string, self int64) (int64, bool) { return C2CPeer(convID, self) }

// IsGroup 是否为群聊 conv_id。
func IsGroup(convID string) bool {
	_, ok := ParseGroupID(convID)
	return ok
}

// IsC2C 是否为单聊 conv_id。
func IsC2C(convID string) bool {
	_, _, ok := ParseC2C(convID)
	return ok
}

func trimC2CPrefix(convID string) (string, bool) {
	for _, p := range []string{PrefixC2C, "d_"} {
		if strings.HasPrefix(convID, p) {
			return strings.TrimPrefix(convID, p), true
		}
	}
	return "", false
}
