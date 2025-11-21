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

// 查询组织同步状态
func GetOrgSyncStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrgSyncStatusRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sync.NewGetOrgSyncStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetOrgSyncStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
