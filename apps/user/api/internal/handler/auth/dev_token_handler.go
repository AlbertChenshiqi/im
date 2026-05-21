package auth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"im/apps/user/api/internal/logic/auth"
	"im/apps/user/api/internal/svc"
	"im/apps/user/api/internal/types"
	"im/pkg/code"
)

func DevTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DevTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			code.WriteHTTP(w, r, err)
			return
		}
		l := auth.NewDevTokenLogic(r.Context(), svcCtx)
		resp, err := l.DevToken(&req)
		if err != nil {
			code.WriteHTTP(w, r, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
