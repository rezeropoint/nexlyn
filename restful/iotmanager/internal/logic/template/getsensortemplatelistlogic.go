package template

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetSensorTemplateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取设备模板列表
func NewGetSensorTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSensorTemplateListLogic {
	return &GetSensorTemplateListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSensorTemplateListLogic) GetSensorTemplateList(req *types.GetSensorTemplateListRequest) (resp *types.GetSensorTemplateListResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_template"),
		logx.Field("operation", "list_templates"),
		logx.Field("status", "started"),
		logx.Field("page", req.Current),
		logx.Field("page_size", req.PageSize),
	).Info("开始获取设备模板列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "list_templates"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetSensorTemplateListResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查模板列表查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTTemplate, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "list_templates"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetSensorTemplateListResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "list_templates"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetSensorTemplateListResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要模板查看权限"},
		}, nil
	}

	// 规范化分页参数
	current, pageSize, _ := svc.NormalizePagination(req.Current, req.PageSize)

	// 构建查询条件
	query := core.TemplateQuery{
		TenantID: jwtUser.TenantId,
		Category: req.Category,
		Keyword:  req.Name, // 名称模糊查询
		Page:     int(current),
		PageSize: int(pageSize),
	}

	// 处理Enabled筛选参数（字符串转bool）
	if req.Enabled != "" {
		enabled := req.Enabled == "true"
		query.Enabled = &enabled
	}

	// 调用IoT引擎查询模板列表
	summaries, total, err := l.svcCtx.IoTEngine.ListTemplates(l.ctx, query)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "list_templates"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("IoT引擎查询模板列表失败")

		return &types.GetSensorTemplateListResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 转换为API类型
	list := svc.ConvertCoreTemplateSummariesToTypes(summaries)

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_template"),
		logx.Field("operation", "list_templates"),
		logx.Field("status", "success"),
		logx.Field("total", total),
		logx.Field("page", current),
		logx.Field("page_size", pageSize),
	).Info("获取设备模板列表成功")

	return &types.GetSensorTemplateListResponse{
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
			List []types.SensorTemplate `json:"list"`
		}{
			List: list,
		},
	}, nil
}
