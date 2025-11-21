// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package infoatomtype

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/logic/infoatomtype"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/lynxmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取信息原子类型详情
func GetInfoAtomTypeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetInfoAtomTypeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := infoatomtype.NewGetInfoAtomTypeLogic(r.Context(), svcCtx)
		resp, err := l.GetInfoAtomType(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
