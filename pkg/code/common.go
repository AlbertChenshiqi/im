package code

// Common 1–999：跨服务通用

const (
	OK Code = 0

	CommonInternal     Code = 1
	CommonInvalidParam Code = 2
	CommonNotFound     Code = 3
	CommonForbidden    Code = 4
	CommonUnauthorized Code = 5
)

func init() {
	register(OK, "ok", "ok")
	register(CommonInternal, "internal", "internal error")
	register(CommonInvalidParam, "invalid_param", "invalid parameter")
	register(CommonNotFound, "not_found", "not found")
	register(CommonForbidden, "forbidden", "forbidden")
	register(CommonUnauthorized, "unauthorized", "unauthorized")
}
