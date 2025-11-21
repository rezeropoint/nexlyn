package orgmapping

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/logic/orgmapping"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取单个组织映射
func GetOrgMappingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrgMappingRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := orgmapping.NewGetOrgMappingLogic(r.Context(), svcCtx)
		resp, err := l.GetOrgMapping(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
