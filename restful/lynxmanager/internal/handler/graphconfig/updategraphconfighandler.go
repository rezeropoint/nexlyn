// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package graphconfig

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/logic/graphconfig"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新逻辑图配置
func UpdateGraphConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateGraphConfigRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := graphconfig.NewUpdateGraphConfigLogic(r.Context(), svcCtx)
		resp, err := l.UpdateGraphConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
