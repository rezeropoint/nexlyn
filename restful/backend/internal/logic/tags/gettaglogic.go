package tags

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取标签详情
func NewGetTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTagLogic {
	return &GetTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTagLogic) GetTag(req *types.GetTagRequest) (resp *types.GetTagResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "get_tag"),
		logx.Field("status", "started"),
		logx.Field("tag_id", req.Id),
	).Info("开始获取标签详情")

	// 获取当前操作者（用于权限验证）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户失败")
		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 查询标签详情（包含租户信息用于权限检查）
	var tagRow struct {
		Id          string    `db:"id"`
		Scope       string    `db:"scope"`
		Label       string    `db:"label"`
		Description string    `db:"description"`
		TenantId    *string   `db:"tenant_id"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	querySQL := `
		SELECT id, scope, label, 
		       COALESCE(description, '') as description,
		       tenant_id, created_at, updated_at
		FROM system_tag_definitions 
		WHERE id = $1
	`

	err = l.svcCtx.DBConn.QueryRowPartial(&tagRow, querySQL, req.Id)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("tag_id", req.Id),
			logx.Field("error", err.Error()),
		).Error("标签不存在")
		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 404, Msg: "标签不存在"},
		}, nil
	}

	// 租户隔离检查：用户标签只能被同租户用户查看
	if tagRow.Scope == "user" && tagRow.TenantId != nil {
		if *tagRow.TenantId != jwtUser.TenantId {
			return &types.GetTagResponse{
				BaseResponse: types.BaseResponse{Code: 403, Msg: "无权查看其他租户的用户标签"},
			}, nil
		}
	}

	// 权限验证：根据scope类型检查读取权限
	var permissionResource core.Resource
	if tagRow.Scope == "tenant" {
		permissionResource = core.ResourceTagTenant
	} else {
		permissionResource = core.ResourceTagUser
	}

	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(jwtUser.UserKey, jwtUser.TenantKey, core.Permission{Resource: permissionResource, Action: core.ActionRead})
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")
		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "tags"),
			logx.Field("operation", "get_tag"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")
		return &types.GetTagResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: fmt.Sprintf("权限不足，需要%s读取权限", permissionResource)},
		}, nil
	}

	// 记录操作成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "tags"),
		logx.Field("operation", "get_tag"),
		logx.Field("status", "success"),
		logx.Field("tag_id", req.Id),
		logx.Field("user_key", jwtUser.UserKey),
	).Info("获取标签详情成功")

	// 构建返回数据
	tagDefinition := types.TagDefinition{
		Id:          tagRow.Id,
		Scope:       tagRow.Scope,
		Label:       tagRow.Label,
		Description: tagRow.Description,
		CreatedAt:   tagRow.CreatedAt.Unix(),
		UpdatedAt:   tagRow.UpdatedAt.Unix(),
	}

	return &types.GetTagResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "获取成功"},
		Data:         tagDefinition,
	}, nil
}
