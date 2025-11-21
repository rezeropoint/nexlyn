package tags

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建标签
func NewCreateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTagLogic {
	return &CreateTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTagLogic) CreateTag(req *types.CreateTagRequest) (resp *types.CreateTagResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "create_tag"),
		logx.Field("status", "started"),
	).Info("开始创建标签")

	// 验证scope值
	if req.Scope != "tenant" && req.Scope != "user" {
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "scope参数必须为tenant或user"},
		}, nil
	}

	// 标准化label（去除前后空格）
	req.Label = strings.TrimSpace(req.Label)
	if req.Label == "" {
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "label参数不能为空"},
		}, nil
	}

	// 标准化description
	if req.Description != "" {
		req.Description = strings.TrimSpace(req.Description)
	}

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "create_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：根据scope类型检查不同权限
	var permissionResource core.Resource
	if req.Scope == "tenant" {
		permissionResource = core.ResourceTagTenant
	} else {
		permissionResource = core.ResourceTagUser
	}

	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: permissionResource, Action: core.ActionWrite})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "create_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "create_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: fmt.Sprintf("权限不足，需要%s管理权限", permissionResource)},
		}, nil
	}

	// 检查是否已存在相同scope和label的标签（考虑租户隔离）
	var existingCount int
	var checkSQL string
	var checkArgs []interface{}

	if req.Scope == "tenant" {
		// 租户标签全局唯一（tenant_id为NULL）
		checkSQL = "SELECT COUNT(*) FROM system_tag_definitions WHERE scope = $1 AND label = $2 AND tenant_id IS NULL"
		checkArgs = []interface{}{req.Scope, req.Label}
	} else {
		// 用户标签在租户内唯一
		checkSQL = "SELECT COUNT(*) FROM system_tag_definitions WHERE scope = $1 AND label = $2 AND tenant_id = $3"
		checkArgs = []interface{}{req.Scope, req.Label, jwtUser.TenantId}
	}

	err = l.svcCtx.DBConn.QueryRow(&existingCount, checkSQL, checkArgs...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "create_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("检查标签重复性失败")
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "检查标签重复性失败"},
		}, nil
	}

	if existingCount > 0 {
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 409, Msg: fmt.Sprintf("该scope下已存在名为'%s'的标签", req.Label)},
		}, nil
	}

	// 生成新的标签ID
	tagId := uuid.New().String()
	now := time.Now().Unix()

	// 插入标签记录（根据scope类型决定tenant_id）
	var insertSQL string
	var insertArgs []interface{}

	if req.Scope == "tenant" {
		// 租户标签，tenant_id为NULL（全局）
		insertSQL = `
			INSERT INTO system_tag_definitions (id, scope, label, description, tenant_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NULL, to_timestamp($5), to_timestamp($6))
		`
		insertArgs = []interface{}{tagId, req.Scope, req.Label, req.Description, now, now}
	} else {
		// 用户标签，指定租户ID
		insertSQL = `
			INSERT INTO system_tag_definitions (id, scope, label, description, tenant_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, to_timestamp($6), to_timestamp($7))
		`
		insertArgs = []interface{}{tagId, req.Scope, req.Label, req.Description, jwtUser.TenantId, now, now}
	}

	_, err = l.svcCtx.DBConn.Exec(insertSQL, insertArgs...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "create_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", tagId),
			logx.Field("error", err.Error()),
		).Error("插入标签记录失败")
		return &types.CreateTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "创建标签失败"},
		}, nil
	}

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "create_tag"),
		logx.Field("status", "success"),
		logx.Field("tag_id", tagId),
		logx.Field("scope", req.Scope),
		logx.Field("label", req.Label),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("标签创建成功")

	return &types.CreateTagResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "创建成功"},
		Data: struct {
			Id string `json:"id"`
		}{Id: tagId},
	}, nil
}
