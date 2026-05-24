package sessionid

import "im/pkg/convid"

// FromConvID 会话排序/分片用 sessionId（与 conv_id 一致：c2c_小_大 或 group_群ID）。
func FromConvID(convID string) string {
	return convID
}

// C2C 单聊 sessionId。
func C2C(a, b int64) string {
	return convid.C2C(a, b)
}

// Group 群聊 sessionId。
func Group(groupID int64) string {
	return convid.Group(groupID)
}
