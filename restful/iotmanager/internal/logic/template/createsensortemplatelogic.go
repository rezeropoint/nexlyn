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

type CreateSensorTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建设备模板
func NewCreateSensorTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSensorTemplateLogic {
	return &CreateSensorTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSensorTemplateLogic) CreateSensorTemplate(req *types.CreateSensorTemplateRequest) (resp *types.CreateSensorTemplateResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_template"),
		logx.Field("operation", "create_template"),
		logx.Field("status", "started"),
		logx.Field("model", req.Model),
		logx.Field("name", req.Name),
	).Info("开始创建设备模板")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "create_template"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.CreateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查模板创建权限
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
			logx.Field("operation", "create_template"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.CreateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "create_template"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.CreateSensorTemplateResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要模板管理权限"},
		}, nil
	}

	// 构建模板元数据
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

	// 转换配置（如果提供）
	var onlineConfig *core.OnlineDetectionConfig
	var businessConfig *core.DataProcessingConfig

	// 检查并转换在线检测配置
	if req.OnlineConfig.Category != "" && req.OnlineConfig.Model != "" && req.OnlineConfig.TopicSuffix != "" {
		onlineConfig = svc.ConvertOnlineDetectionConfigToCore(req.OnlineConfig, jwtUser.TenantId)
	}

	// 检查并转换业务数据处理配置
	if req.BusinessConfig.Category != "" && req.BusinessConfig.Model != "" && req.BusinessConfig.TopicSuffix != "" {
		businessConfig = svc.ConvertDataProcessingConfigToCore(req.BusinessConfig, jwtUser.TenantId)
	}

	// 检查并转换控制配置
	var controlConfig *core.DeviceControlConfig
	if req.ControlConfig.Category != "" && req.ControlConfig.Model != "" && req.ControlConfig.CommandSuffix != "" {
		controlConfig = svc.ConvertDeviceControlConfigToCore(req.ControlConfig, jwtUser.TenantId)
	}

	// 调用IoT引擎创建模板
	templateID, err := l.svcCtx.IoTEngine.CreateTemplate(l.ctx, metadata, onlineConfig, businessConfig, controlConfig)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_template"),
			logx.Field("operation", "create_template"),
			logx.Field("status", "failed"),
			logx.Field("model", req.Model),
			logx.Field("error", err.Error()),
		).Error("IoT引擎创建模板失败")

		return &types.CreateSensorTemplateResponse{
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
		logx.Field("operation", "create_template"),
		logx.Field("status", "success"),
		logx.Field("template_id", templateID),
		logx.Field("model", req.Model),
		logx.Field("name", req.Name),
	).Info("创建设备模板成功")

	return &types.CreateSensorTemplateResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Id    string `json:"id"`
			Model string `json:"model"`
			Name  string `json:"name"`
		}{
			Id:    templateID,
			Model: req.Model,
			Name:  req.Name,
		},
	}, nil
}
