/**
 * 时序数据查询API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { BaseResponse } from "../types/common";
import type {
  DeviceLatestValuesItem,
  DeviceStatisticsItem,
  TimeSeriesDataItem,
} from "../types/timeseries";

// 查询时序数据请求
export interface QueryTimeSeriesRequest {
  orgId: string; // 组织ID（必填）
  deviceIds?: string[]; // 设备ID列表（可选）
  fieldNames: string[]; // 查询字段列表（必填）
  startTime: string; // 开始时间（ISO8601格式）
  endTime: string; // 结束时间（ISO8601格式）
  aggregation?: string; // 聚合类型（avg/max/min/sum/count/last）
  interval?: number; // 聚合时间间隔（秒）
  limit?: number; // 限制返回条数
  offset?: number; // 偏移量
  orderBy?: string; // 排序字段
  orderDir?: string; // 排序方向
}

// 查询时序数据响应
export interface QueryTimeSeriesResponse extends BaseResponse {
  data?: {
    list: TimeSeriesDataItem[];
    total: number;
    page: number;
    pageSize: number;
  };
}

export async function queryTimeSeries(
  data: QueryTimeSeriesRequest,
  options?: { [key: string]: any }
) {
  return request<QueryTimeSeriesResponse>(`${API_PREFIX}/timeseries/query`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
    ...(options || {}),
  });
}

// 获取最新值请求
export interface GetLatestValuesRequest {
  orgId: string; // 组织ID（必填）
  deviceIds?: string[]; // 设备ID列表（可选）
  fieldNames: string[]; // 字段列表（必填）
}

// 获取最新值响应
export interface GetLatestValuesResponse extends BaseResponse {
  data?: DeviceLatestValuesItem[];
}

export async function getLatestValues(
  data: GetLatestValuesRequest,
  options?: { [key: string]: any }
) {
  return request<GetLatestValuesResponse>(
    `${API_PREFIX}/timeseries/latest-values`,
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

// 获取设备统计请求
export interface GetDeviceStatisticsRequest {
  orgId: string; // 组织ID（必填）
  deviceIds?: string[]; // 设备ID列表（可选）
  fieldNames: string[]; // 字段列表（必填）
  startTime: string; // 开始时间（ISO8601格式）
  endTime: string; // 结束时间（ISO8601格式）
}

// 获取设备统计响应
export interface GetDeviceStatisticsResponse extends BaseResponse {
  data?: DeviceStatisticsItem[];
}

export async function getDeviceStatistics(
  data: GetDeviceStatisticsRequest,
  options?: { [key: string]: any }
) {
  return request<GetDeviceStatisticsResponse>(
    `${API_PREFIX}/timeseries/statistics`,
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
