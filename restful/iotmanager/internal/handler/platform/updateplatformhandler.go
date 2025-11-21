package platform

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/logic/platform"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新平台配置
func UpdatePlatformHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdatePlatformRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := platform.NewUpdatePlatformLogic(r.Context(), svcCtx)
		resp, err := l.UpdatePlatform(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
