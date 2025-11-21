/**
 * 设备类别和标准字段API
 */
import { request } from "@umijs/max";
import { API_PREFIX } from "../constants";
import type { DeviceCategoryInfo, StandardFieldInfo } from "../types/category";
import type { BaseResponse } from "../types/common";

// 获取所有设备类别
export interface GetDeviceCategoriesResponse extends BaseResponse {
  data?: {
    categories: DeviceCategoryInfo[];
  };
}

export async function getDeviceCategories(options?: { [key: string]: any }) {
  return request<GetDeviceCategoriesResponse>(
    `${API_PREFIX}/device-categories`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

// 获取指定类别的标准字段列表
export interface GetStandardFieldsResponse extends BaseResponse {
  data?: {
    fields: StandardFieldInfo[];
  };
}

export async function getStandardFields(
  category: string,
  options?: { [key: string]: any }
) {
  return request<GetStandardFieldsResponse>(
    `${API_PREFIX}/device-categories/${category}/standard-fields`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}
