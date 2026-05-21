package code

import (
	"errors"
	"fmt"
)

// Code 业务错误码（整型，按服务分段）
type Code int

// Meta 错误元数据
type Meta struct {
	Slug    string // 对外短码（如 WebSocket error.code）
	Message string // 默认提示
}

var registry = make(map[Code]Meta)

func register(c Code, slug, message string) Code {
	registry[c] = Meta{Slug: slug, Message: message}
	return c
}

func (c Code) Int() int { return int(c) }

func (c Code) Slug() string {
	if m, ok := registry[c]; ok && m.Slug != "" {
		return m.Slug
	}
	return fmt.Sprintf("%d", c)
}

func (c Code) Message() string {
	if m, ok := registry[c]; ok {
		return m.Message
	}
	return "error"
}

// Error 带分段的业务错误
type Error struct {
	Code Code
	Msg  string
}

func (e *Error) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Code.Message()
}

// New 构造业务错误，可选覆盖默认文案
func New(c Code, msg ...string) *Error {
	e := &Error{Code: c}
	if len(msg) > 0 && msg[0] != "" {
		e.Msg = msg[0]
	}
	return e
}

// Is 判断是否为指定业务错误码
func Is(err error, c Code) bool {
	e, ok := As(err)
	return ok && e.Code == c
}

// As 提取 *Error
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// HTTPStatus 根据错误码建议 HTTP 状态
func HTTPStatus(c Code) int {
	switch c {
	case CommonUnauthorized, UserDevAuthDisabled:
		return 401
	case CommonForbidden:
		return 403
	case CommonNotFound, UserNotFound:
		return 404
	case UserLoginNotReady:
		return 501
	default:
		return 400
	}
}
