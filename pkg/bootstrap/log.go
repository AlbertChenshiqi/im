package bootstrap

import (
	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/logx"
)

// SilenceZeroNoise 关闭 go-zero 框架 stat（usage.go）与 load shedding（sheddingstat.go）周期日志。
// 应在 conf.MustLoad 之后调用；配置中建议同时设置 Log.Stat: false。
func SilenceZeroNoise() {
	logx.DisableStat()
	load.DisableLog()
}
