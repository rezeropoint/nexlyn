// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package userassignments

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/eventhandler/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserAssignmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户任务列表（待办、已处理、抄送）
func NewGetUserAssignmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAssignmentsLogic {
	return &GetUserAssignmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserAssignmentsLogic) GetUserAssignments(req *types.GetUserAssignmentsRequest) (resp *types.GetUserAssignmentsResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user_assignments"),
		logx.Field("operation", "get_user_assignments"),
		logx.Field("status", "started"),
		logx.Field("category", req.Category),
		logx.Field("page", req.Page),
		logx.Field("page_size", req.PageSize),
	).Info("开始获取用户任务列表")

	// 1. 从JWT获取用户信息（无需额外权限验证，用户只能查自己的任务）
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user_assignments"),
			logx.Field("operation", "get_user_assignments"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.GetUserAssignmentsResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 2. 调用Skylark引擎获取用户任务列表
	// SDK会自动处理：
	//   - localUserID到remoteUserID的映射
	//   - 根据tenantID获取平台配置
	//   - FlowID和FlowTitle的批量查询补充（避免N+1问题）
	assignments, total, err := l.svcCtx.SkylarkEngine.GetUserAssignments(
		l.ctx,
		jwtUser.TenantId,
		jwtUser.UserId, // localUserID - SDK会自动转换为remoteUserID
		req.Category,   // proposed/processed/cc
		req.Page,
		req.PageSize,
	)

	if err != nil {
		code, msg := svc.HandleSkylarkError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "user_assignments"),
			logx.Field("operation", "get_user_assignments"),
			logx.Field("status", "failed"),
			logx.Field("tenant_id", jwtUser.TenantId),
			logx.Field("user_id", jwtUser.UserId),
			logx.Field("category", req.Category),
			logx.Field("error", err.Error()),
		).Error("获取用户任务列表失败")

		return &types.GetUserAssignmentsResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 3. 转换结果
	assignmentList := svc.ConvertCoreAssignmentListToTypes(assignments)

	// 4. 记录成功
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "user_assignments"),
		logx.Field("operation", "get_user_assignments"),
		logx.Field("status", "success"),
		logx.Field("tenant_id", jwtUser.TenantId),
		logx.Field("user_id", jwtUser.UserId),
		logx.Field("category", req.Category),
		logx.Field("count", len(assignmentList)),
		logx.Field("total", total),
	).Info("获取用户任务列表成功")

	return &types.GetUserAssignmentsResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: types.AssignmentListData{
			List:  assignmentList,
			Total: total,
		},
	}, nil
}
