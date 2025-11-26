// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package httpreceive

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetHttpReceiveListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 HTTP 接收配置列表
func NewGetHttpReceiveListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHttpReceiveListLogic {
	return &GetHttpReceiveListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHttpReceiveListLogic) GetHttpReceiveList(req *types.GetHttpReceiveListRequest) (resp *types.GetHttpReceiveListResponse, err error) {
	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "list"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetHttpReceiveListResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTHttpReceive, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		return &types.GetHttpReceiveListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		return &types.GetHttpReceiveListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 规范化分页参数
	current, pageSize, _ := svc.NormalizePagination(req.Current, req.PageSize)

	// 构建查询条件
	query := core.HttpReceiveQuery{
		TenantID: jwtUser.TenantId,
		Keyword:  req.Keyword,
		Page:     int(current),
		PageSize: int(pageSize),
	}

	// 调用IoT引擎查询
	list, total, err := l.svcCtx.IoTEngine.ListHttpReceiveConfigs(l.ctx, query)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("module", "iot_http_receive"),
			logx.Field("operation", "list"),
			logx.Field("error", err.Error()),
		).Error("查询HTTP接收配置列表失败")

		return &types.GetHttpReceiveListResponse{
			BaseResponse: types.BaseResponse{Code: code, Msg: msg},
		}, nil
	}

	// 转换为API类型
	metadataList := svc.ConvertCoreHttpReceiveSummaryListToTypes(list)

	return &types.GetHttpReceiveListResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		PageParams: types.PageParams{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
		Data: struct {
			List []types.HttpReceiveMetadata `json:"list"`
		}{
			List: metadataList,
		},
	}, nil
}
