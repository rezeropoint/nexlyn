package permission

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/backend/internal/logic/permission"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 为用户移除角色
func RemoveRoleForUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoveRoleRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := permission.NewRemoveRoleForUserLogic(r.Context(), svcCtx)
		resp, err := l.RemoveRoleForUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
