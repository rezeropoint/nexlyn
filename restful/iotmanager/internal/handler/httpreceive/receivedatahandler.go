// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package httpreceive

import (
	"fmt"
	"net/http"

	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/logic/httpreceive"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

// 接收 HTTP 数据
func ReceiveDataHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 使用 EmiyaRegistry 解析请求体（支持嵌套 JSON）
		var data map[string]any
		if err := svcCtx.EmiyaRegistry.AnalyzingBody(&r.Body, &data); err != nil {
			httpx.ErrorCtx(r.Context(), w, fmt.Errorf("解析请求体失败: %v", err))
			logx.WithContext(r.Context()).WithFields(
				logx.Field("service", svcCtx.Config.RestConf.Name),
				logx.Field("pod", svcCtx.PodName),
				logx.Field("module", "http_receive"),
				logx.Field("operation", "receive_data"),
				logx.Field("status", "failed"),
				logx.Field("error", err.Error()),
			).Error("解析请求体失败")
			return
		}

		// 2. 调用 Logic 层
		l := httpreceive.NewReceiveDataLogic(r.Context(), svcCtx)
		resp, err := l.ReceiveData(pathvar.Vars(r)["configId"], data)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
