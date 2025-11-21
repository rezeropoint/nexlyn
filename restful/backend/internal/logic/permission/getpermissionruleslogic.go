package permission

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionRulesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取权限规则列表
func NewGetPermissionRulesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionRulesLogic {
	return &GetPermissionRulesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionRulesLogic) GetPermissionRules(req *types.GetPermissionRulesRequest) (resp *types.GetPermissionRulesResponse, err error) {

	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "get_permission_rules"),
		logx.Field("status", "started"),
		logx.Field("user_id", req.UserKey),
		logx.Field("tenant_id", req.TenantKey),
		logx.Field("resource", req.Resource),
		logx.Field("action", req.Action),
		logx.Field("ptype", req.Ptype),
	).Info("开始获取权限规则列表")

	// 从JWT中获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "get_permission_rules"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")

		return &types.GetPermissionRulesResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户认证失败",
			},
		}, nil
	}
	// 权限验证：检查是否有权限管理权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: core.ResourcePermission, Action: core.ActionWrite})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "get_permission_rules"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.GetPermissionRulesResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "系统权限检查失败",
			},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "permission"),
			logx.Field("operation", "get_permission_rules"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", "permission denied"),
		).Error("用户无权限执行此操作")

		return &types.GetPermissionRulesResponse{
			BaseResponse: types.BaseResponse{
				Code: 403,
				Msg:  "权限不足，需要权限管理权限",
			},
		}, nil
	}

	// 构建权限规则列表
	var rules []types.PermissionRule

	// 根据不同场景获取权限规则
	if req.UserKey != "" {
		// 有用户指定的情况
		if req.Ptype == "" || req.Ptype == "p" {
			// 获取用户直接权限
			permissions, err := l.svcCtx.Casbinx.GetDirectPermissionsSecure(jwtUser.UserKey, req.UserKey, req.TenantKey)
			if err != nil {
				logx.WithContext(l.ctx).Error("获取用户权限失败: " + err.Error())
				return &types.GetPermissionRulesResponse{
					BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
				}, nil
			}
			for _, perm := range permissions {
				rules = append(rules, types.PermissionRule{
					Ptype: "p", UserKey: req.UserKey, TenantKey: req.TenantKey,
					Resource: string(perm.Resource), Action: string(perm.Action),
				})
			}
		}

		if req.Ptype == "" || req.Ptype == "g" {
			// 获取用户角色分配
			roles, err := l.svcCtx.Casbinx.GetUserRoles(req.UserKey, req.TenantKey)
			if err != nil {
				logx.WithContext(l.ctx).Error("获取用户角色失败: " + err.Error())
				return &types.GetPermissionRulesResponse{
					BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
				}, nil
			}
			for _, role := range roles {
				rules = append(rules, types.PermissionRule{
					Ptype: "g", UserKey: req.UserKey, RoleKey: role, TenantKey: req.TenantKey,
				})
			}
		}
	} else if req.Ptype == "p" {
		// 无用户指定，但要求直接权限 - CasbinX接口暂不支持GetAllDirectPolicies方法
		// 类似GetAllGroupingPolicies，需要扩展接口添加GetAllDirectPolicies(tenantKey string) ([]DirectPolicy, error)
		logx.WithContext(l.ctx).Info("请求获取租户所有直接权限，但CasbinX接口暂不支持此功能")
	} else if req.Ptype == "g" {
		// 无用户指定，但要求角色分配
		allGroupings, err := l.svcCtx.Casbinx.GetAllGroupingPolicies(req.TenantKey)
		if err != nil {
			logx.WithContext(l.ctx).Error("获取角色分配失败: " + err.Error())
			return &types.GetPermissionRulesResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "获取角色分配失败"},
			}, nil
		}
		for _, grouping := range allGroupings {
			rules = append(rules, types.PermissionRule{
				Ptype: "g", UserKey: grouping.UserKey, RoleKey: grouping.RoleKey, TenantKey: grouping.TenantKey,
			})
		}
	}

	// 设置分页信息
	total := int64(len(rules))
	pageParams := types.PageParams{
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    total,
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "permission"),
		logx.Field("operation", "get_permission_rules"),
		logx.Field("status", "success"),
		logx.Field("total_count", total),
		logx.Field("operator_user_id", jwtUser.UserId),
	).Info("获取权限规则列表成功")

	return &types.GetPermissionRulesResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "获取权限规则列表成功",
		},
		Data:       rules,
		PageParams: pageParams,
	}, nil
}
