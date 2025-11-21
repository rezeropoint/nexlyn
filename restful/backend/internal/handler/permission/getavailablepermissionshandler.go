package permission

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/backend/internal/logic/permission"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取可用权限列表
func GetAvailablePermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := permission.NewGetAvailablePermissionsLogic(r.Context(), svcCtx)
		resp, err := l.GetAvailablePermissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
