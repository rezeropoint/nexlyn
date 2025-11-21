// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flowjourney

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/logic/flowjourney"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取当前处理人
func GetCurrentProcessingUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCurrentProcessingUsersRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flowjourney.NewGetCurrentProcessingUsersLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentProcessingUsers(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
