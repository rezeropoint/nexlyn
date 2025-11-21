// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package control

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/logic/control"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 查询AI Box任务列表
func ListAIBoxTasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAIBoxTasksRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := control.NewListAIBoxTasksLogic(r.Context(), svcCtx)
		resp, err := l.ListAIBoxTasks(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
