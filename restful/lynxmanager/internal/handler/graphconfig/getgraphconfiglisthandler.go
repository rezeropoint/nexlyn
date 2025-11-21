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

// 查询逻辑图配置列表
func GetGraphConfigListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetGraphConfigListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := graphconfig.NewGetGraphConfigListLogic(r.Context(), svcCtx)
		resp, err := l.GetGraphConfigList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
