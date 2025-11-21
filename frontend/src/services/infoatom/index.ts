// @ts-ignore
/* eslint-disable */
import { request } from "@umijs/max";

const API_PREFIX = "/api/v1";

export async function getInfoAtomTypeList(
  params: API.GetInfoAtomTypeListRequest,
  options?: { [key: string]: any }
) {
  return request<API.GetInfoAtomTypeListResponse>(
    `${API_PREFIX}/infoatom/type/list`,
    {
      method: "GET",
      params,
      ...(options || {}),
    }
  );
}

export async function getInfoAtomType(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.GetInfoAtomTypeResponse>(
    `${API_PREFIX}/infoatom/type/${params.id}`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

export async function checkInfoAtomType(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.CheckInfoAtomTypeResponse>(
    `${API_PREFIX}/infoatom/type/${params.id}/check`,
    {
      method: "GET",
      ...(options || {}),
    }
  );
}

export async function createInfoAtomType(
  body: API.CreateInfoAtomTypeRequest,
  options?: { [key: string]: any }
) {
  return request<API.CreateInfoAtomTypeResponse>(
    `${API_PREFIX}/infoatom/type`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      data: body,
      ...(options || {}),
    }
  );
}

export async function updateInfoAtomType(
  params: { id: string },
  body: API.UpdateInfoAtomTypeRequest,
  options?: { [key: string]: any }
) {
  return request<API.UpdateInfoAtomTypeResponse>(
    `${API_PREFIX}/infoatom/type/${params.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      data: body,
      ...(options || {}),
    }
  );
}

export async function deleteInfoAtomType(
  params: { id: string },
  options?: { [key: string]: any }
) {
  return request<API.DeleteInfoAtomTypeResponse>(
    `${API_PREFIX}/infoatom/type/${params.id}`,
    {
      method: "DELETE",
      ...(options || {}),
    }
  );
}
