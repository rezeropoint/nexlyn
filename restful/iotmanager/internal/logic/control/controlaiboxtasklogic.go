// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package control

import (
	"context"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/iotmanager/internal/types"

	casbinxcore "github.com/rezeropoint/casbinx/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type ControlAIBoxTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 控制AI Box任务（启动/停止）
func NewControlAIBoxTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ControlAIBoxTaskLogic {
	return &ControlAIBoxTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ControlAIBoxTaskLogic) ControlAIBoxTask(req *types.ControlAIBoxTaskRequest) (resp *types.ControlAIBoxTaskResponse, err error) {
	// 记录操作开始
	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device_control"),
		logx.Field("operation", "control_aibox_task"),
		logx.Field("status", "started"),
		logx.Field("device_id", req.DeviceId),
		logx.Field("task_id", req.TaskId),
		logx.Field("control_command", req.ControlCommand),
	).Info("开始控制AI Box任务")

	// 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device_control"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("获取JWT用户信息失败")

		return &types.ControlAIBoxTaskResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权: " + err.Error(),
			},
		}, nil
	}

	// 权限验证：检查设备操作权限
	hasPermission, err := l.svcCtx.Casbinx.CheckPermission(
		jwtUser.UserKey,
		jwtUser.TenantKey,
		casbinxcore.Permission{Resource: auth.ResourceIoTDevice, Action: casbinxcore.ActionWrite},
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device_control"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
			logx.Field("error", err.Error()),
		).Error("权限检查失败")

		return &types.ControlAIBoxTaskResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "系统权限检查失败"},
		}, nil
	}

	if !hasPermission {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device_control"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("status", "failed"),
			logx.Field("user_key", jwtUser.UserKey),
		).Error("用户权限不足")

		return &types.ControlAIBoxTaskResponse{
			BaseResponse: types.BaseResponse{Code: 403, Msg: "权限不足：需要设备操作权限"},
		}, nil
	}

	// 查询用户组织权限范围（包括所有子组织）
	orgIds, err := auth.GetAllChildOrgIds(l.ctx, l.svcCtx.DBConn, req.OrgId, jwtUser.TenantId)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device_control"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("status", "failed"),
			logx.Field("error", err.Error()),
		).Error("查询组织权限范围失败")

		return &types.ControlAIBoxTaskResponse{
			BaseResponse: types.BaseResponse{Code: 500, Msg: "查询组织失败: " + err.Error()},
		}, nil
	}

	// 调用IoT引擎控制AI Box任务
	err = l.svcCtx.IoTEngine.ControlAIBoxTask(l.ctx, jwtUser.TenantId, req.DeviceId, orgIds, req.TaskId, req.ControlCommand)
	if err != nil {
		code, msg := svc.HandleIoTError(err)
		logx.WithContext(l.ctx).WithFields(
			logx.Field("service", l.svcCtx.Config.RestConf.Name),
			logx.Field("pod", l.svcCtx.PodName),
			logx.Field("module", "iot_device_control"),
			logx.Field("operation", "control_aibox_task"),
			logx.Field("status", "failed"),
			logx.Field("device_id", req.DeviceId),
			logx.Field("task_id", req.TaskId),
			logx.Field("control_command", req.ControlCommand),
			logx.Field("error", err.Error()),
		).Error("IoT引擎控制AI Box任务失败")

		return &types.ControlAIBoxTaskResponse{
			BaseResponse: types.BaseResponse{
				Code: code,
				Msg:  msg,
			},
		}, nil
	}

	// 记录成功
	commandDesc := "停止"
	if req.ControlCommand == 1 {
		commandDesc = "启动"
	}

	logx.WithContext(l.ctx).WithFields(
		logx.Field("service", l.svcCtx.Config.RestConf.Name),
		logx.Field("pod", l.svcCtx.PodName),
		logx.Field("module", "iot_device_control"),
		logx.Field("operation", "control_aibox_task"),
		logx.Field("status", "success"),
		logx.Field("device_id", req.DeviceId),
		logx.Field("task_id", req.TaskId),
		logx.Field("control_command", req.ControlCommand),
		logx.Field("command_desc", commandDesc),
	).Info("控制AI Box任务成功")

	return &types.ControlAIBoxTaskResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "success",
		},
		Data: struct {
			Message string `json:"message"`
		}{
			Message: "任务" + commandDesc + "命令已发送",
		},
	}, nil
}
