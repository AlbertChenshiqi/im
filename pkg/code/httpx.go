package code

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type httpBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// WriteHTTP 写入 REST 错误 JSON：{"code":10101,"msg":"..."}
func WriteHTTP(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := As(err); ok {
		msg := e.Msg
		if msg == "" {
			msg = e.Code.Message()
		}
		httpx.WriteJsonCtx(r.Context(), w, HTTPStatus(e.Code), httpBody{
			Code: e.Code.Int(),
			Msg:  msg,
		})
		return
	}
	httpx.ErrorCtx(r.Context(), w, err)
}
