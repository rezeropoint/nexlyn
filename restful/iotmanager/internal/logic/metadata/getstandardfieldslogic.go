package metadata

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetStandardFieldsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定类别的标准字段列表
func NewGetStandardFieldsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStandardFieldsLogic {
	return &GetStandardFieldsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetStandardFieldsLogic) GetStandardFields(req *types.GetStandardFieldsRequest) (resp *types.GetStandardFieldsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_metadata"),
		logx.Field("operation", "get_standard_fields"),
		logx.Field("status", "started"),
		logx.Field("category", req.Category),
	).Info("开始获取标准字段列表")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_metadata"),
			logx.Field("operation", "get_standard_fields"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetStandardFieldsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查元数据查看权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTMetadata, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_metadata"),
			logx.Field("operation", "get_standard_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetStandardFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_metadata"),
			logx.Field("operation", "get_standard_fields"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.GetStandardFieldsResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要元数据查看权限"},
		}, nil
	}

	// 从IoT引擎获取标准字段列表
	fields := l.svcCtx.IoTEngine.GetStandardFields(core.DeviceCategory(req.Category))

	// 转换为响应类型
	respFields := make([]types.StandardFieldInfo, 0, len(fields))
	for _, field := range fields {
		var minValue, maxValue float64
		if field.MinValue != nil {
			minValue = *field.MinValue
		}
		if field.MaxValue != nil {
			maxValue = *field.MaxValue
		}

		respFields = append(respFields, types.StandardFieldInfo{
			Name:        field.Name,
			DisplayName: field.DisplayName,
			FieldType:   string(field.FieldType),
			Unit:        field.Unit,
			Description: field.Description,
			Required:    field.Required,
			MinValue:    minValue,
			MaxValue:    maxValue,
			EnumValues:  field.EnumValues,
		})
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_metadata"),
		logx.Field("operation", "get_standard_fields"),
		logx.Field("status", "success"),
		logx.Field("category", req.Category),
		logx.Field("count", len(respFields)),
	).Info("获取标准字段列表成功")

	return &types.GetStandardFieldsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Fields []types.StandardFieldInfo `json:"fields"`
		}{
			Fields: respFields,
		},
	}, nil
}
