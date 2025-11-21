// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package sync

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/backend/internal/logic/sync"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 绑定组织到已存在的Skylark组织
func BindOrganizationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindOrganizationRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sync.NewBindOrganizationLogic(r.Context(), svcCtx)
		resp, err := l.BindOrganization(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
