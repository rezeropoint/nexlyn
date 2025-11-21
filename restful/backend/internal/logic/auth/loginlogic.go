package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/config"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {

	// 记录登录开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "auth"),
		logx.Field("operation", "login"),
		logx.Field("status", "started"),
		logx.Field("username", req.UserName),
	).Info("开始用户登录")

	// 查询用户信息（用户名已经是全局唯一的）
	var user struct {
		Id           string `db:"id"`
		UserKey      string `db:"user_key"`
		UserName     string `db:"user_name"`
		PasswordHash string `db:"password_hash"`
		Name         string `db:"name"`
		Email        string `db:"email"`
		TenantID     string `db:"tenant_id"`

		Status string `db:"status"`
		// 租户信息
		TenantName       string `db:"tenant_name"`
		TenantIdStr      string `db:"tenant_id_str"`     // 租户UUID（用于数据库关联）
		TenantIdentifier string `db:"tenant_identifier"` // 租户标识符（用于Casbin）

		// 主组织信息（使用sql.NullString防止NULL值报错）
		PrimaryOrgId   sql.NullString `db:"primary_org_id"`   // 主组织UUID
		PrimaryOrgName sql.NullString `db:"primary_org_name"` // 主组织名称
	}

	query := `
		SELECT u.id, u.user_key, u.user_name, u.password_hash, u.name, u.email,
		       u.tenant_id, u.status,
		       t.tenant_name, u.tenant_id as tenant_id_str, t.tenant_key as tenant_identifier,
		       org.id as primary_org_id, org.name as primary_org_name
		FROM system_users u
		LEFT JOIN system_tenants t ON u.tenant_id = t.id
		LEFT JOIN system_user_org_relations uor ON u.id = uor.user_id AND uor.is_primary = true AND uor.status = 'active' AND uor.deleted_at IS NULL
		LEFT JOIN system_organizations org ON uor.org_id = org.id AND org.status = 'active' AND org.deleted_at IS NULL
		WHERE u.user_name = $1 AND u.status = 'active'
	`

	err = l.svcCtx.DBConn.QueryRowCtx(l.ctx, &user, query, req.UserName)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "auth"),
			logx.Field("operation", "login"),
			logx.Field("status", "failed"),
			logx.Field("username", req.UserName),
			logx.Field("error", err.Error()),
		).Error("用户不存在或已被禁用")

		return &types.LoginResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户名或密码错误",
			},
			Status: "error",
		}, nil
	}

	// 添加调试日志
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "auth"),
		logx.Field("operation", "login"),
		logx.Field("status", "debug"),
		logx.Field("username", req.UserName),
		logx.Field("password_hash", user.PasswordHash),
		logx.Field("input_password", req.Password),
	).Debug("密码验证调试信息")

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "auth"),
			logx.Field("operation", "login"),
			logx.Field("status", "failed"),
			logx.Field("username", req.UserName),
			logx.Field("error", err.Error()),
		).Error("密码验证失败")

		return &types.LoginResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "用户名或密码错误",
			},
			Status: "error",
		}, nil
	}

	// 提取主组织信息（处理可能为NULL的情况）
	primaryOrgId := ""
	primaryOrgName := ""
	if user.PrimaryOrgId.Valid {
		primaryOrgId = user.PrimaryOrgId.String
	}
	if user.PrimaryOrgName.Valid {
		primaryOrgName = user.PrimaryOrgName.String
	}

	// 生成JWT令牌（同时包含租户key和UUID）
	jwtUser := auth.CreateJWTUser(
		user.UserKey, user.Id, user.UserName,
		user.TenantIdentifier, user.TenantIdStr, user.TenantName,
		primaryOrgId, primaryOrgName,
		user.Status,
	)

	// 添加调试日志
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "auth"),
		logx.Field("operation", "login"),
		logx.Field("status", "debug"),
		logx.Field("username", req.UserName),
		logx.Field("userKey", user.UserKey),
		logx.Field("userId", user.Id),
		logx.Field("tenantIdentifier", user.TenantIdentifier),
		logx.Field("tenantIdStr", user.TenantIdStr),
		logx.Field("tenantName", user.TenantName),
		logx.Field("jwtUser", jwtUser),
	).Info("创建JWT用户信息")

	payload, err := json.Marshal(jwtUser)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "auth"),
			logx.Field("operation", "login"),
			logx.Field("status", "failed"),
			logx.Field("username", req.UserName),
			logx.Field("error", err.Error()),
		).Error("序列化用户信息失败")

		return &types.LoginResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgAuthLogin, err),
			},
			Status: "error",
		}, nil
	}

	token, err := auth.GetJwtToken(l.svcCtx.Config.Auth.AccessSecret, time.Now().Unix(), l.svcCtx.Config.Auth.AccessExpire, string(payload))
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "auth"),
			logx.Field("operation", "login"),
			logx.Field("status", "failed"),
			logx.Field("username", req.UserName),
			logx.Field("error", err.Error()),
		).Error("生成JWT token失败")

		return &types.LoginResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  config.FormatError(config.ErrMsgAuthToken, err),
			},
			Status: "error",
		}, nil
	}

	// 更新用户最后登录时间
	_, err = l.svcCtx.DBConn.Exec(`
		UPDATE system_users 
		SET last_login_at = CURRENT_TIMESTAMP
		WHERE user_name = $1 AND tenant_id = $2
	`, req.UserName, user.TenantID)
	if err != nil {
		// 记录错误但不影响登录
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "auth"),
			logx.Field("operation", "login"),
			logx.Field("status", "warning"),
			logx.Field("username", req.UserName),
			logx.Field("error", err.Error()),
		).Error("更新最后登录时间失败")
	}

	// 登录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "auth"),
		logx.Field("operation", "login"),
		logx.Field("status", "success"),
		logx.Field("username", req.UserName),
	).Info("用户登录成功")

	return &types.LoginResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "登录成功",
		},
		Status: "ok",
		Type:   "account",
		Token:  token,
	}, nil
}
