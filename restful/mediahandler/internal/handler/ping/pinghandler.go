package ping

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/mediahandler/internal/logic/ping"
	"github.com/rezeropoint/nexlyn/restful/mediahandler/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ping.NewPingLogic(r.Context(), svcCtx)
		err := l.Ping()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
