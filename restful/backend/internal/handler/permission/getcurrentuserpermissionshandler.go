package permission

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/backend/internal/logic/permission"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取当前用户权限
func GetCurrentUserPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := permission.NewGetCurrentUserPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentUserPermissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
