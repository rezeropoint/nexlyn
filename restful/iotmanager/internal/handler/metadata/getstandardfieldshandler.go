package metadata

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/logic/metadata"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取指定类别的标准字段列表
func GetStandardFieldsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetStandardFieldsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := metadata.NewGetStandardFieldsLogic(r.Context(), svcCtx)
		resp, err := l.GetStandardFields(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
