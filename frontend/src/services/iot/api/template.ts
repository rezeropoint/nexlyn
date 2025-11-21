/**
 * 设备模板管理API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse, PageParams } from "../types/common";
import type {
  DataProcessingConfig,
  DeviceControlConfig,
  OnlineDetectionConfig,
  SensorTemplate,
  SensorTemplateDetail,
} from "../types/template";

// 获取模板列表请求
export interface GetSensorTemplateListRequest {
  page?: number;
  pageSize?: number;
  model?: string; // 型号筛选
  name?: string; // 名称模糊查询
  category?: string; // 类别筛选
  manufacturer?: string; // 厂商筛选
  enabled?: string; // 启用状态筛选（true/false）
}

// 获取模板列表响应
export interface GetSensorTemplateListResponse
  extends BaseResponse,
    PageParams {
  data?: {
    list: SensorTemplate[];
  };
}

export async function getSensorTemplateList(
  params?: GetSensorTemplateListRequest,
  options?: { [key: string]: any }
) {
  return request<GetSensorTemplateListResponse>(
    `${API_PREFIX}/sensor-template/list`,
    {
      method: "GET",
      params,
      ...(options || {}),
    }
  );
}

// 获取模板详情
export interface GetSensorTemplateResponse extends BaseResponse {
  data?: SensorTemplateDetail;
}

export async function getSensorTemplate(
  id: string,
  options?: { [key: string]: any }
) {
  return request<GetSensorTemplateResponse>(
    `${API_PREFIX}/sensor-template/${id}`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

// 创建模板请求
export interface CreateSensorTemplateRequest {
  model: string; // 设备型号标识（必填，唯一）
  name: string; // 模板显示名称（必填）
  category?: string; // 设备类别
  manufacturer?: string; // 设备厂商
  description?: string; // 模板描述
  version?: string; // 模板版本号
  enabled?: boolean; // 是否启用（默认true）
  onlineConfig?: OnlineDetectionConfig; // 在线检测配置（可选）
  businessConfig?: DataProcessingConfig; // 业务数据处理配置（可选）
  controlConfig?: DeviceControlConfig; // 设备控制配置（可选）
}

// 创建模板响应
export interface CreateSensorTemplateResponse extends BaseResponse {
  data?: {
    id: string;
    model: string;
    name: string;
  };
}

export async function createSensorTemplate(
  data: CreateSensorTemplateRequest,
  options?: { [key: string]: any }
) {
  return request<CreateSensorTemplateResponse>(
    `${API_PREFIX}/sensor-template`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      data,
      ...(options || {}),
    }
  );
}

// 更新模板请求
// 注意：category 和 model 是模板核心标识，前端必须传递但不允许修改
export interface UpdateSensorTemplateRequest {
  model: string; // 设备型号（必填，不允许修改）
  name?: string;
  category: string; // 设备类别（必填，不允许修改）
  manufacturer?: string;
  description?: string;
  version?: string;
  enabled?: boolean;
  onlineConfig?: OnlineDetectionConfig; // 在线检测配置（可选，传null表示删除）
  businessConfig?: DataProcessingConfig; // 业务数据处理配置（可选，传null表示删除）
  controlConfig?: DeviceControlConfig; // 设备控制配置（可选，传null表示删除）
}

export async function updateSensorTemplate(
  id: string,
  data: UpdateSensorTemplateRequest,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/sensor-template/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 删除模板
export async function deleteSensorTemplate(
  id: string,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/sensor-template/${id}`, {
    method: "DELETE",
    ...(options || {}),
  });
}
