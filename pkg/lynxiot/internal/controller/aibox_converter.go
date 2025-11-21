package controller

import (
	"encoding/json"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core/devices"
)

// aiboxConverter AI Box型号转换器
// 实现真实设备协议与统一接口的转换
type aiboxConverter struct{}

// newAIBoxConverter 创建AI Box型号转换器实例
func newAIBoxConverter() *aiboxConverter {
	return &aiboxConverter{}
}

// ConvertCommand 将统一命令转换为真实设备协议的MQTT消息
func (c *aiboxConverter) ConvertCommand(model string, command interface{}) ([]byte, error) {
	cmd, ok := command.(*devices.AIBoxControlCommand)
	if !ok {
		return nil, fmt.Errorf("无效的命令类型，期望 *devices.AIBoxControlCommand")
	}

	// 根据命令类型转换
	switch cmd.CommandType {
	case devices.AIBoxCommandGetCapabilities:
		// 获取算法能力
		realCmd := devices.AIBoxRealCommand{
			BoardId: cmd.DeviceID,
			Event:   devices.AIBoxEventAbilityFetch,
		}
		return json.Marshal(realCmd)

	case devices.AIBoxCommandListTasks:
		// 查询任务列表
		realCmd := devices.AIBoxRealCommand{
			BoardId: cmd.DeviceID,
			Event:   devices.AIBoxEventTaskFetch,
		}
		return json.Marshal(realCmd)

	case devices.AIBoxCommandControlTask:
		// 控制任务（启动/停止）
		payload, ok := cmd.Payload.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("ControlTask命令的Payload必须是map[string]interface{}")
		}

		taskID, ok := payload["task_id"].(string)
		if !ok {
			return nil, fmt.Errorf("ControlTask命令缺少task_id参数")
		}

		controlCommand, ok := payload["control_command"].(int)
		if !ok {
			return nil, fmt.Errorf("ControlTask命令缺少control_command参数")
		}

		realCmd := devices.AIBoxTaskControlCommand{
			BoardId:        cmd.DeviceID,
			Event:          devices.AIBoxEventTaskControl,
			ControlCommand: controlCommand,
			AlgTaskSession: taskID,
		}
		return json.Marshal(realCmd)

	case devices.AIBoxCommandCreateTask,
		devices.AIBoxCommandUpdateTask,
		devices.AIBoxCommandDeleteTask:
		// 暂未实现的功能
		return nil, core.ErrNotImplemented

	default:
		return nil, fmt.Errorf("未知的命令类型: %s", cmd.CommandType)
	}
}

// ConvertResponse 将真实设备响应转换为统一格式
func (c *aiboxConverter) ConvertResponse(deviceID string, response []byte) (*devices.AIBoxControlResponse, error) {
	// 先检查是哪种响应类型（通过Event字段判断）
	var baseResp struct {
		Event  string              `json:"Event"`
		Result devices.AIBoxResult `json:"Result"`
	}
	if err := json.Unmarshal(response, &baseResp); err != nil {
		return nil, fmt.Errorf("解析响应基础信息失败: %w", err)
	}

	// 检查Result.Code（0成功，非0失败）
	if baseResp.Result.Code != 0 {
		return &devices.AIBoxControlResponse{
			DeviceID: deviceID,
			Event:    baseResp.Event,
			Status:   devices.AIBoxStatusFailed,
			Error:    fmt.Sprintf("设备返回错误: %s (Code: %d)", baseResp.Result.Desc, baseResp.Result.Code),
		}, nil
	}

	// 根据Event类型解析不同的响应
	switch baseResp.Event {
	case devices.AIBoxEventAbilityFetch:
		// 算法能力查询响应
		return c.convertAbilityResponse(deviceID, response)

	case devices.AIBoxEventTaskFetch:
		// 任务列表查询响应
		return c.convertTaskListResponse(deviceID, response)

	case devices.AIBoxEventTaskControl:
		// 任务控制响应
		return c.convertTaskControlResponse(deviceID, response)

	default:
		return nil, fmt.Errorf("未知的响应类型: %s", baseResp.Event)
	}
}

// convertAbilityResponse 转换算法能力查询响应
func (c *aiboxConverter) convertAbilityResponse(deviceID string, response []byte) (*devices.AIBoxControlResponse, error) {
	var abilityResp devices.AIBoxAbilityResponse
	if err := json.Unmarshal(response, &abilityResp); err != nil {
		return nil, fmt.Errorf("解析算法能力响应失败: %w", err)
	}

	// 直接使用真实协议结构
	capabilities := &devices.AIBoxCapabilities{
		BoardId:   abilityResp.BoardId,
		Abilities: abilityResp.Ability,
	}

	return &devices.AIBoxControlResponse{
		DeviceID: deviceID,
		Event:    devices.AIBoxEventAbilityFetch,
		Status:   devices.AIBoxStatusSuccess,
		Result:   capabilities,
	}, nil
}

// convertTaskListResponse 转换任务列表查询响应
func (c *aiboxConverter) convertTaskListResponse(deviceID string, response []byte) (*devices.AIBoxControlResponse, error) {
	var taskResp devices.AIBoxTaskResponse
	if err := json.Unmarshal(response, &taskResp); err != nil {
		return nil, fmt.Errorf("解析任务列表响应失败: %w", err)
	}

	// 直接映射真实协议字段，无需map转换
	tasks := make([]devices.AIBoxAlgorithmTask, 0, len(taskResp.Content))
	for _, taskInfo := range taskResp.Content {
		task := devices.AIBoxAlgorithmTask{
			AlgTaskSession: taskInfo.AlgTaskSession,
			TaskDesc:       taskInfo.TaskDesc,
			MediaName:      taskInfo.MediaName,
			AlgInfo:        taskInfo.AlgInfo,
			AlgTaskStatus:  taskInfo.AlgTaskStatus,
			AlarmBody:      taskInfo.AlarmBody,
			AlarmProtocol:  taskInfo.AlarmProtocol,
			MetadataUrl:    taskInfo.MetadataUrl,
			UserData:       taskInfo.UserData,
		}
		tasks = append(tasks, task)
	}

	return &devices.AIBoxControlResponse{
		DeviceID: deviceID,
		Event:    devices.AIBoxEventTaskFetch,
		Status:   devices.AIBoxStatusSuccess,
		Result:   tasks,
	}, nil
}

// convertTaskControlResponse 转换任务控制响应
func (c *aiboxConverter) convertTaskControlResponse(deviceID string, response []byte) (*devices.AIBoxControlResponse, error) {
	var controlResp devices.AIBoxTaskControlResponse
	if err := json.Unmarshal(response, &controlResp); err != nil {
		return nil, fmt.Errorf("解析任务控制响应失败: %w", err)
	}

	// 构建响应结果（包含任务标识和描述信息）
	result := map[string]interface{}{
		"alg_task_session": controlResp.AlgTaskSession,
		"description":      controlResp.Result.Desc,
	}

	return &devices.AIBoxControlResponse{
		DeviceID: deviceID,
		Event:    devices.AIBoxEventTaskControl,
		Status:   devices.AIBoxStatusSuccess,
		Result:   result,
	}, nil
}
