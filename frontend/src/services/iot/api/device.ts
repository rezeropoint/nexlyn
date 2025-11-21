/**
 * 设备绑定管理API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse, PageParams } from "../types/common";
import type {
  DeviceBinding,
  DeviceBindingSummary,
  UnboundDevice,
} from "../types/device";

// 绑定设备请求
export interface BindDeviceRequest {
  deviceId: string; // 设备唯一标识符(必填)
  deviceName: string; // 设备名称(必填)
  deviceAlias?: string; // 设备别名
  deviceModel: string; // 设备型号(必填,关联模板)
  description?: string; // 设备描述
  location?: string; // 安装位置
  installationDate?: string; // 安装日期
  status?: string; // 设备状态(默认active)
  tagIds?: string[]; // 标签ID列表(绑定时)
  orgId: string; // 组织ID(必填)
}

// 绑定设备响应
export interface BindDeviceResponse extends BaseResponse {
  data?: {
    id: string; // 设备绑定记录ID
    deviceId: string; // 设备ID
  };
}

export async function bindDevice(
  data: BindDeviceRequest,
  options?: { [key: string]: any }
) {
  return request<BindDeviceResponse>(`${API_PREFIX}/device-bindings`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 获取设备详情
export interface GetDeviceResponse extends BaseResponse {
  data?: DeviceBinding;
}

export async function getDevice(
  deviceId: string,
  orgId: string,
  options?: { [key: string]: any }
) {
  return request<GetDeviceResponse>(
    `${API_PREFIX}/device-bindings/${deviceId}`,
    {
      method: "GET",
      params: { orgId },
      ...(options || {}),
    }
  );
}

// 获取设备列表请求
export interface ListDevicesRequest {
  page?: number;
  pageSize?: number;
  orgId: string; // 组织ID筛选（必填）
  deviceModel?: string; // 设备型号筛选
  deviceCategory?: string; // 设备类别筛选
  status?: string; // 设备状态筛选
  isOnline?: string; // 在线状态筛选(true/false)
  keyword?: string; // 关键词搜索(设备ID、名称、别名)
}

// 获取设备列表响应
export interface ListDevicesResponse extends BaseResponse, PageParams {
  data?: {
    list: DeviceBindingSummary[];
  };
}

export async function listDevices(
  params?: ListDevicesRequest,
  options?: { [key: string]: any }
) {
  return request<ListDevicesResponse>(`${API_PREFIX}/device-bindings/list`, {
    method: "GET",
    params,
    ...(options || {}),
  });
}

// 获取未绑定设备列表
export interface ListUnboundDevicesResponse extends BaseResponse {
  data?: {
    list: UnboundDevice[];
  };
}

export async function listUnboundDevices(options?: { [key: string]: any }) {
  return request<ListUnboundDevicesResponse>(
    `${API_PREFIX}/device-bindings/unbound`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

// 更新设备请求
export interface UpdateDeviceRequest {
  orgId: string; // 组织ID（必填）
  deviceName?: string; // 设备名称
  deviceAlias?: string; // 设备别名
  description?: string; // 设备描述
  location?: string; // 安装位置
  installationDate?: string; // 安装日期
  status?: string; // 设备状态
  tagIds?: string[]; // 更新标签列表(完全替换)
}

export async function updateDevice(
  deviceId: string,
  orgId: string,
  data: UpdateDeviceRequest,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/device-bindings/${deviceId}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data: { ...data, orgId },
    ...(options || {}),
  });
}

// 解绑设备
export async function unbindDevice(
  deviceId: string,
  orgId: string,
  options?: { [key: string]: any }
) {
  return request<BaseResponse>(`${API_PREFIX}/device-bindings/${deviceId}`, {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
    data: { orgId },
    ...(options || {}),
  });
}
