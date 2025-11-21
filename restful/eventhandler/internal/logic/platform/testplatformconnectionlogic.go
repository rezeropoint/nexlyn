package platform

import (
	"context"

	"github.com/rezeropoint/go-skylark/v2/core"
	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type TestPlatformConnectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 测试Skylark平台连接
func NewTestPlatformConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestPlatformConnectionLogic {
	return &TestPlatformConnectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestPlatformConnectionLogic) TestPlatformConnection(req *types.TestPlatformConnectionRequest) (resp *types.TestPlatformConnectionResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "test_platform_connection"),
		logx.Field("status", "started"),
	).Info("开始测试Skylark平台连接")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "test_platform_connection"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.TestPlatformConnectionResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查平台配置读取权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceSkylarkPlatform, Action: casbinxcore.ActionRead},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "test_platform_connection"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.TestPlatformConnectionResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "test_platform_connection"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.TestPlatformConnectionResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要Skylark平台配置查看权限"},
		}, nil
	}

	// 构建测试用的平台配置
	testCfg := &core.PlatformConfig{
		Host:        req.Host,
		Port:        req.Port,
		Database:    req.Database,
		Username:    req.Username,
		Password:    req.Password,
		NamespaceID: req.NamespaceId,
		APIBaseURL:  req.ApiBaseUrl, // 新增字段
		APIToken:    req.ApiToken,   // 新增字段
	}

	// 调用Skylark引擎验证连接
	err = l.svcCtx.SkylarkEngine.ValidatePlatformConfig(l.ctx, testCfg)

	if err != nil {
		// 连接失败
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "skylark_platform"),
			logx.Field("operation", "test_platform_connection"),
			logx.Field("status", "failed"),
			logx.Field("host", req.Host),
			logx.Field("error", err.Error()),
		).Error("Skylark平台连接测试失败")

		return &types.TestPlatformConnectionResponse{
			BaseResponse: types.BaseResponse{
				Code: 0,
				Msg:  "success",
			},
			Data: struct {
				Success    bool   `json:"success"`
				FlowsCount int    `json:"flowsCount,optional"`
				Message    string `json:"message,optional"`
			}{
				Success: false,
				Message: err.Error(),
			},
		}, nil
	}

	// 连接成功，尝试获取flows数量
	flows, err := l.svcCtx.SkylarkEngine.GetFlowList(l.ctx, jwtUser.TenantId)
	flowsCount := 0
	if err == nil {
		flowsCount = len(flows)
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "skylark_platform"),
		logx.Field("operation", "test_platform_connection"),
		logx.Field("status", "success"),
		logx.Field("host", req.Host),
		logx.Field("flows_count", flowsCount),
	).Info("Skylark平台连接测试成功")

	return &types.TestPlatformConnectionResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Success    bool   `json:"success"`
			FlowsCount int    `json:"flowsCount,optional"`
			Message    string `json:"message,optional"`
		}{
			Success:    true,
			FlowsCount: flowsCount,
			Message:    "连接成功",
		},
	}, nil
}
