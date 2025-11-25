// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flowjourney

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/logic/flowjourney"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取流程完整详情（一站式接口，包含基础信息、审批历史、待处理节点）
func GetJourneyFullDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetJourneyFullDetailRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flowjourney.NewGetJourneyFullDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetJourneyFullDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
