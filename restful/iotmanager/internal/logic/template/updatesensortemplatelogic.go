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

type UpdateSensorTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新设备模板
func NewUpdateSensorTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSensorTemplateLogic {
	return &UpdateSensorTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSensorTemplateLogic) UpdateSensorTemplate(req *types.UpdateSensorTemplateRequest) (resp *types.UpdateSensorTemplateResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_template"),
		logx.Field("operation", "update_template"),
		logx.Field("status", "started"),
		logx.Field("template_id", req.Id),
	).Info("开始更新设备模板")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "update_template"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.UpdateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查模板更新权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTTemplate, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "update_template"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.UpdateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "update_template"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.UpdateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要模板管理权限"},
		}, nil
	}

	// 构建模板元数据
	// 注意：Model和Category必须传递（前端从getsensortemplatelogic.go获取），后端会验证不允许修改
	metadata := core.TemplateMetadata{
		Model:        req.Model,
		Name:         req.Name,
		Category:     req.Category,
		Manufacturer: req.Manufacturer,
		Description:  req.Description,
		Version:      req.Version,
		Enabled:      req.Enabled,
		TenantID:     jwtUser.TenantId,
		CreatedBy:    jwtUser.UserId,
	}

	// 转换配置（如果提供了）
	var onlineConfig *core.OnlineDetectionConfig
	var businessConfig *core.DataProcessingConfig

	// 检查并转换在线检测配置
	if req.OnlineConfig != nil && req.OnlineConfig.Category != "" && req.OnlineConfig.Model != "" && req.OnlineConfig.TopicSuffix != "" {
		onlineConfig = svc.ConvertOnlineDetectionConfigToCore(*req.OnlineConfig, jwtUser.TenantId)
	}

	// 检查并转换业务数据处理配置
	if req.BusinessConfig != nil && req.BusinessConfig.Category != "" && req.BusinessConfig.Model != "" && req.BusinessConfig.TopicSuffix != "" {
		businessConfig = svc.ConvertDataProcessingConfigToCore(*req.BusinessConfig, jwtUser.TenantId)
	}

	// 检查并转换控制配置
	var controlConfig *core.DeviceControlConfig
	if req.ControlConfig != nil && req.ControlConfig.Category != "" && req.ControlConfig.Model != "" && req.ControlConfig.CommandSuffix != "" {
		controlConfig = svc.ConvertDeviceControlConfigToCore(*req.ControlConfig, jwtUser.TenantId)
	}

	// 调用IoT引擎更新模板
	err = l.svcCtx.IoTEngine.UpdateTemplate(l.ctx, req.Id, metadata, onlineConfig, businessConfig, controlConfig)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "update_template"),
			logx.Field("status", "failed"),
			logx.Field("template_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("IoT引擎更新模板失败")

		return &types.UpdateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_template"),
		logx.Field("operation", "update_template"),
		logx.Field("status", "success"),
		logx.Field("template_id", req.Id),
	).Info("更新设备模板成功")

	return &types.UpdateSensorTemplateResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
	}, nil
}
