package query

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/logic/query"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 查询时序数据
func QueryTimeSeriesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryTimeSeriesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := query.NewQueryTimeSeriesLogic(r.Context(), svcCtx)
		resp, err := l.QueryTimeSeries(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
