package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"im/apps/user/api/internal/logic/user"
	"im/apps/user/api/internal/svc"
	"im/pkg/code"
)

func SetOnlineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewSetOnlineLogic(r.Context(), svcCtx)
		resp, err := l.SetOnline()
		if err != nil {
			code.WriteHTTP(w, r, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
