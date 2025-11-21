package tags

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除标签
func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTagLogic) DeleteTag(req *types.DeleteTagRequest) (resp *types.DeleteTagResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "delete_tag"),
		logx.Field("status", "started"),
		logx.Field("tag_id", req.Id),
	).Info("开始删除标签")

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 先检查标签是否存在，并获取其信息用于权限验证和租户隔离
	var existingTag struct {
		Scope    string  `db:"scope"`
		Label    string  `db:"label"`
		TenantId *string `db:"tenant_id"`
	}

	checkSQL := "SELECT scope, label, tenant_id FROM system_tag_definitions WHERE id = $1"
	err = l.svcCtx.DBConn.QueryRowPartial(&existingTag, checkSQL, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("标签不存在")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "标签不存在"},
		}, nil
	}

	// 租户隔离检查：用户标签只能被同租户用户删除
	if existingTag.Scope == "user" && existingTag.TenantId != nil {
		if *existingTag.TenantId != jwtUser.TenantId {
			return &types.DeleteTagResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "无权删除其他租户的用户标签"},
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

	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: permissionResource, Action: core.ActionDelete})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: fmt.Sprintf("权限不足，需要%s删除权限", permissionResource)},
		}, nil
	}

	// 检查是否有用户或租户正在使用该标签
	var userTagCount int
	userTagSQL := "SELECT COUNT(*) FROM user_tag_relations WHERE tag_id = $1"
	err = l.svcCtx.DBConn.QueryRow(&userTagCount, userTagSQL, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("检查用户标签关联失败")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "检查标签使用情况失败"},
		}, nil
	}

	var tenantTagCount int
	tenantTagSQL := "SELECT COUNT(*) FROM system_tenant_tag_relations WHERE tag_id = $1"
	err = l.svcCtx.DBConn.QueryRow(&tenantTagCount, tenantTagSQL, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("检查租户标签关联失败")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "检查标签使用情况失败"},
		}, nil
	}

	// 如果有用户或租户正在使用该标签，不允许删除
	if userTagCount > 0 || tenantTagCount > 0 {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("user_tag_count", userTagCount),
			logx.Field("tenant_tag_count", tenantTagCount),
		).Info("标签正在被使用，不能删除")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 409, Msg: "标签正在被使用，不能删除。请先移除所有关联关系后再试"},
		}, nil
	}

	// 执行删除操作
	deleteSQL := "DELETE FROM system_tag_definitions WHERE id = $1"
	_, err = l.svcCtx.DBConn.Exec(deleteSQL, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "delete_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("删除标签失败")
		return &types.DeleteTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "删除标签失败"},
		}, nil
	}

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "delete_tag"),
		logx.Field("status", "success"),
		logx.Field("tag_id", req.Id),
		logx.Field("tag_label", existingTag.Label),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("标签删除成功")

	return &types.DeleteTagResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "删除成功"},
	}, nil
}
