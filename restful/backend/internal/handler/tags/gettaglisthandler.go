package tags

import (
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/backend/internal/logic/tags"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取标签列表
func GetTagListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTagListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tags.NewGetTagListLogic(r.Context(), svcCtx)
		resp, err := l.GetTagList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
