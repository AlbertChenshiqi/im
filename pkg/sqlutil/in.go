package sqlutil

import "strings"

// Placeholders 生成 n 个 `?` 占位符，用于 IN 子句。
func Placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(",?", n)[1:]
}

// Int64Args 将 []int64 转为 []any 供 Exec/Query 使用。
func Int64Args(ids []int64) []any {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}
