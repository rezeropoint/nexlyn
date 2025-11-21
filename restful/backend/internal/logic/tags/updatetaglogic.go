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

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新标签
func NewUpdateTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTagLogic {
	return &UpdateTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTagLogic) UpdateTag(req *types.UpdateTagRequest) (resp *types.UpdateTagResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "update_tag"),
		logx.Field("status", "started"),
		logx.Field("tag_id", req.Id),
	).Info("开始更新标签")

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "update_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.UpdateTagResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 先检查标签是否存在，并获取其信息用于权限验证
	var existingTag struct {
		Scope       string  `db:"scope"`
		Label       string  `db:"label"`
		Description string  `db:"description"`
		TenantId    *string `db:"tenant_id"`
	}

	checkSQL := "SELECT scope, label, description, tenant_id FROM system_tag_definitions WHERE id = $1"
	err = l.svcCtx.DBConn.QueryRowPartial(&existingTag, checkSQL, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "update_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("标签不存在")
		return &types.UpdateTagResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "标签不存在"},
		}, nil
	}

	// 租户隔离检查：用户标签只能被同租户用户修改
	if existingTag.Scope == "user" && existingTag.TenantId != nil {
		if *existingTag.TenantId != jwtUser.TenantId {
			return &types.UpdateTagResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "无权修改其他租户的用户标签"},
			}, nil
		}
	}

	// 权限验证：根据scope类型检查不同权限
	var permissionResource core.Resource
	if existingTag.Scope == "tenant" {
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
			logx.Field("operation", "update_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.UpdateTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "update_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.UpdateTagResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: fmt.Sprintf("权限不足，需要%s管理权限", permissionResource)},
		}, nil
	}

	// 准备更新字段
	updateFields := []string{}
	updateArgs := []interface{}{}
	argIndex := 1

	// 处理scope更新
	if req.Scope != "" {
		if req.Scope != "tenant" && req.Scope != "user" {
			return &types.UpdateTagResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "scope参数必须为tenant或user"},
			}, nil
		}
		updateFields = append(updateFields, fmt.Sprintf("scope = $%d", argIndex))
		updateArgs = append(updateArgs, req.Scope)
		argIndex++
	}

	// 处理label更新
	if req.Label != "" {
		label := strings.TrimSpace(req.Label)
		if label == "" {
			return &types.UpdateTagResponse{
				BaseResponse: types.BaseResponse{Code: 400, Msg: "label参数不能为空"},
			}, nil
		}

		// 检查同一scope下是否已存在相同label的其他标签（考虑租户隔离）
		checkScope := req.Scope
		if checkScope == "" {
			checkScope = existingTag.Scope
		}

		var conflictCount int
		var conflictSQL string
		var conflictArgs []interface{}

		if checkScope == "tenant" {
			// 租户标签全局唯一检查
			conflictSQL = "SELECT COUNT(*) FROM system_tag_definitions WHERE scope = $1 AND label = $2 AND tenant_id IS NULL AND id != $3"
			conflictArgs = []interface{}{checkScope, label, req.Id}
		} else {
			// 用户标签在租户内唯一检查
			conflictSQL = "SELECT COUNT(*) FROM system_tag_definitions WHERE scope = $1 AND label = $2 AND tenant_id = $3 AND id != $4"
			conflictArgs = []interface{}{checkScope, label, jwtUser.TenantId, req.Id}
		}

		err = l.svcCtx.DBConn.QueryRow(&conflictCount, conflictSQL, conflictArgs...)
		if err != nil {
			logx.WithContext(l.ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "tags"),
				logx.Field("operation", "update_tag"),
				logx.Field("status", "failed"),
				logx.Field("error", err.Error()),
			).Error("检查标签重复性失败")
			return &types.UpdateTagResponse{
				BaseResponse: types.BaseResponse{Code: 500, Msg: "检查标签重复性失败"},
			}, nil
		}

		if conflictCount > 0 {
			return &types.UpdateTagResponse{
				BaseResponse: types.BaseResponse{Code: 409, Msg: fmt.Sprintf("该scope下已存在名为'%s'的其他标签", label)},
			}, nil
		}

		updateFields = append(updateFields, fmt.Sprintf("label = $%d", argIndex))
		updateArgs = append(updateArgs, label)
		argIndex++
	}

	// 处理description更新
	if req.Description != "" {
		description := strings.TrimSpace(req.Description)
		updateFields = append(updateFields, fmt.Sprintf("description = $%d", argIndex))
		updateArgs = append(updateArgs, description)
		argIndex++
	}

	// 如果没有任何更新字段，返回错误
	if len(updateFields) == 0 {
		return &types.UpdateTagResponse{
			BaseResponse: types.BaseResponse{Code: 400, Msg: "没有提供任何更新字段"},
		}, nil
	}

	// 添加updated_at字段
	now := time.Now().Unix()
	updateFields = append(updateFields, fmt.Sprintf("updated_at = to_timestamp($%d)", argIndex))
	updateArgs = append(updateArgs, now)
	argIndex++

	// 添加WHERE条件参数
	updateArgs = append(updateArgs, req.Id)

	// 构建并执行更新SQL
	updateSQL := fmt.Sprintf("UPDATE system_tag_definitions SET %s WHERE id = $%d", strings.Join(updateFields, ", "), argIndex)

	_, err = l.svcCtx.DBConn.Exec(updateSQL, updateArgs...)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "update_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("更新标签失败")
		return &types.UpdateTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "更新标签失败"},
		}, nil
	}

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "update_tag"),
		logx.Field("status", "success"),
		logx.Field("tag_id", req.Id),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("标签更新成功")

	return &types.UpdateTagResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "更新成功"},
	}, nil
}
