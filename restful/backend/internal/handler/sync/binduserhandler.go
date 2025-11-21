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

// 绑定用户到已存在的Skylark用户
func BindUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindUserRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sync.NewBindUserLogic(r.Context(), svcCtx)
		resp, err := l.BindUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
