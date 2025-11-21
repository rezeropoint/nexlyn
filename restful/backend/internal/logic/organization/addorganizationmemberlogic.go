package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"
	pb "github.com/rezeropoint/nexlyn/service/eventsync/pb"

	"github.com/rezeropoint/casbinx/core"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AddOrganizationMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 添加组织成员
func NewAddOrganizationMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOrganizationMemberLogic {
	return &AddOrganizationMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddOrganizationMemberLogic) AddOrganizationMember(req *types.AddOrganizationMemberRequest) (resp *types.AddOrganizationMemberResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "add_organization_member"),
		logx.Field("status", "started"),
		logx.Field("org_id", req.Id),
		logx.Field("user_id", req.UserId),
		logx.Field("is_primary", req.IsPrimary),
	).Info("开始添加组织成员")

	// 获取当前用户信息
	currentUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "add_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("从 JWT 获取用户信息失败")

		return &types.AddOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 401, Msg: "用户认证失败"},
		}, nil
	}

	// 权限验证：检查组织写入权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		currentUser.UserKey,
		currentUser.TenantKey,
		core.Permission{Resource: core.ResourceOrganization, Action: core.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "add_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.AddOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgPermissionCheck, err)},
		}, nil
	}
	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "add_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
		).Error("权限不足")

		return &types.AddOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足"},
		}, nil
	}

	// 获取租户ID
	var tenantId string
	getTenantQuery := `SELECT id FROM system_tenants WHERE tenant_key = $1 AND status = 'active'`
	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &tenantId, getTenantQuery, currentUser.TenantKey)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "add_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("error", err.Error()),
		).Error("获取租户信息失败")

		return &types.AddOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: config.FormatError(config.ErrMsgTenantQuery, err)},
		}, nil
	}

	// 使用事务处理成员添加
	err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 验证组织存在且属于当前租户
		var orgCount int
		checkOrgQuery := `SELECT COUNT(*) FROM system_organizations WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		err := session.QueryRowCtx(ctx, &orgCount, checkOrgQuery, req.Id, tenantId)
		if err != nil || orgCount == 0 {
			return fmt.Errorf(config.ErrMsgOrganizationNotFound)
		}

		// 验证用户存在且属于当前租户
		var userCount int
		checkUserQuery := `SELECT COUNT(*) FROM system_users WHERE id = $1 AND tenant_id = $2 AND status = 'active'`
		err = session.QueryRowCtx(ctx, &userCount, checkUserQuery, req.UserId, tenantId)
		if err != nil || userCount == 0 {
			return fmt.Errorf(config.ErrMsgUserNotFound)
		}

		// 检查是否已存在关系
		var relationCount int
		checkRelationQuery := `SELECT COUNT(*) FROM system_user_org_relations WHERE user_id = $1 AND org_id = $2 AND deleted_at IS NULL`
		err = session.QueryRowCtx(ctx, &relationCount, checkRelationQuery, req.UserId, req.Id)
		if err != nil {
			return fmt.Errorf(config.ErrMsgDatabaseQuery)
		}
		if relationCount > 0 {
			return fmt.Errorf(config.ErrMsgOrganizationMemberExists)
		}

		// 如果设置为主要关系，需先更新其他关系为非主要
		if req.IsPrimary {
			_, err = session.ExecCtx(ctx, `
				UPDATE system_user_org_relations 
				SET is_primary = false 
				WHERE user_id = $1 AND deleted_at IS NULL
			`, req.UserId)
			if err != nil {
				return fmt.Errorf(config.ErrMsgOrganizationMemberUpdate)
			}
		}

		// 插入新的用户组织关系
		insertQuery := `
			INSERT INTO system_user_org_relations (
				user_id, org_id, position_title, is_primary
			) VALUES ($1, $2, $3, $4)
		`
		_, err = session.ExecCtx(ctx, insertQuery, req.UserId, req.Id, req.PositionTitle, req.IsPrimary)
		if err != nil {
			return fmt.Errorf(config.ErrMsgOrganizationMemberAdd)
		}

		// 同步到 Skylark（事务内，失败则回滚）
		if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
			syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			syncResp, syncErr := l.svcCtx.AdminSyncClient.AddMember(syncCtx, &pb.AddMemberReq{
				TenantId:      tenantId,
				LocalOrgId:    req.Id,
				LocalMemberId: req.UserId,
			})

			if syncErr != nil {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "sync_add_member_to_skylark"),
					logx.Field("status", "failed"),
					logx.Field("org_id", req.Id),
					logx.Field("user_id", req.UserId),
					logx.Field("error", syncErr.Error()),
				).Error("同步添加成员到 Skylark 失败，回滚事务")
				return fmt.Errorf("添加成员到 Skylark 失败: %v", syncErr)
			}

			if !syncResp.Success {
				logx.WithContext(ctx).WithFields(
					logx.Field("service", l.svcCtx.Config.RestConf.Name),
					logx.Field("pod", l.svcCtx.PodName),
					logx.Field("module", "organization"),
					logx.Field("operation", "sync_add_member_to_skylark"),
					logx.Field("status", "failed"),
					logx.Field("org_id", req.Id),
					logx.Field("user_id", req.UserId),
					logx.Field("message", syncResp.Message),
					logx.Field("error_code", syncResp.ErrorCode),
				).Error("同步添加成员到 Skylark 失败，回滚事务")
				return fmt.Errorf("添加成员到 Skylark 失败: %s", syncResp.Message)
			}

			logx.WithContext(ctx).WithFields(
				logx.Field("service", l.svcCtx.Config.RestConf.Name),
				logx.Field("pod", l.svcCtx.PodName),
				logx.Field("module", "organization"),
				logx.Field("operation", "sync_add_member_to_skylark"),
				logx.Field("status", "success"),
				logx.Field("org_id", req.Id),
				logx.Field("user_id", req.UserId),
			).Info("成功同步添加成员到 Skylark")
		}

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "organization"),
			logx.Field("operation", "add_organization_member"),
			logx.Field("status", "failed"),
			logx.Field("user_key", currentUser.UserKey),
			logx.Field("tenant_key", currentUser.TenantKey),
			logx.Field("org_id", req.Id),
			logx.Field("user_id", req.UserId),
			logx.Field("error", err.Error()),
		).Error("添加组织成员失败")

		return &types.AddOrganizationMemberResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: err.Error()},
		}, nil
	}

	// 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "organization"),
		logx.Field("operation", "add_organization_member"),
		logx.Field("status", "success"),
		logx.Field("user_key", currentUser.UserKey),
		logx.Field("tenant_key", currentUser.TenantKey),
		logx.Field("org_id", req.Id),
		logx.Field("user_id", req.UserId),
		logx.Field("position", req.PositionTitle),
		logx.Field("is_primary", req.IsPrimary),
	).Info("添加组织成员成功")

	return &types.AddOrganizationMemberResponse{
		BaseResponse: types.BaseResponse{Code: 0, Msg: "添加成功"},
	}, nil
}
