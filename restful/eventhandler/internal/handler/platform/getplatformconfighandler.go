package platform

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/logic/platform"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取Skylark平台配置
func GetPlatformConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPlatformConfigRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := platform.NewGetPlatformConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetPlatformConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
