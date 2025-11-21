package template

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/logic/template"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新设备模板
func UpdateSensorTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateSensorTemplateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := template.NewUpdateSensorTemplateLogic(r.Context(), svcCtx)
		resp, err := l.UpdateSensorTemplate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
