/**
 * AI Box 设备控制API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse } from "../types/common";
import type { AIBoxAlgorithmTask, AIBoxCapabilities } from "../types/control";

// 查询AI Box任务列表请求
export interface ListAIBoxTasksRequest {
  deviceId: string; // 设备ID（路径参数）
  orgId: string; // 组织ID（必填）
}

// 查询AI Box任务列表响应
export interface ListAIBoxTasksResponse extends BaseResponse {
  data?: {
    list: AIBoxAlgorithmTask[];
  };
}

export async function listAIBoxTasks(
  deviceId: string,
  orgId: string,
  options?: { [key: string]: any }
) {
  return request<ListAIBoxTasksResponse>(
    `${API_PREFIX}/device-control/aibox/${deviceId}/tasks`,
    {
      method: "GET",
      params: { orgId },
      ...(options || {}),
    }
  );
}

// 获取AI Box算法能力请求
export interface GetAIBoxCapabilitiesRequest {
  deviceId: string; // 设备ID（路径参数）
  orgId: string; // 组织ID（必填）
}

// 获取AI Box算法能力响应
export interface GetAIBoxCapabilitiesResponse extends BaseResponse {
  data?: AIBoxCapabilities;
}

export async function getAIBoxCapabilities(
  deviceId: string,
  orgId: string,
  options?: { [key: string]: any }
) {
  return request<GetAIBoxCapabilitiesResponse>(
    `${API_PREFIX}/device-control/aibox/${deviceId}/capabilities`,
    {
      method: "GET",
      params: { orgId },
      ...(options || {}),
    }
  );
}

// 控制AI Box任务请求
export interface ControlAIBoxTaskRequest {
  deviceId: string; // 设备ID（路径参数）
  taskId: string; // 任务ID（路径参数）
  orgId: string; // 组织ID（必填）
  controlCommand: number; // 控制命令（0=停止, 1=启动）
}

// 控制AI Box任务响应
export interface ControlAIBoxTaskResponse extends BaseResponse {
  data?: {
    message: string;
  };
}

export async function controlAIBoxTask(
  deviceId: string,
  taskId: string,
  orgId: string,
  controlCommand: number,
  options?: { [key: string]: any }
) {
  return request<ControlAIBoxTaskResponse>(
    `${API_PREFIX}/device-control/aibox/${deviceId}/tasks/${taskId}/control`,
    {
      method: "POST",
      params: { orgId },
      data: { controlCommand },
      ...(options || {}),
    }
  );
}
