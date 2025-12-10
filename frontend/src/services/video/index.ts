import { request } from "@umijs/max";

// ================ 标签管理相关接口 ================

// 设备标签信息
export interface DeviceTag {
  id: string;
  tagName: string;
  description?: string;
  deviceCount: number;
  createdAt: string;
  updatedAt: string;
}

// 标签列表响应
export interface DeviceTagListResponse {
  code: number;
  message: string;
  data: {
    total: number;
    pages: number;
    tags: DeviceTag[];
  };
}

// 创建标签请求
export interface CreateDeviceTagRequest {
  tagName: string;
  description?: string;
}

// 创建标签响应
export interface CreateDeviceTagResponse {
  code: number;
  message: string;
  data: {
    id: string;
    tagName: string;
    description?: string;
    createdAt: string;
    updatedAt: string;
  };
}

// 更新标签请求
export interface UpdateDeviceTagRequest {
  tagName?: string;
  description?: string;
}

// 更新标签响应
export interface UpdateDeviceTagResponse {
  code: number;
  message: string;
  data: {
    id: string;
    tagName: string;
    description?: string;
    updatedAt: string;
  };
}

// 删除标签响应
export interface DeleteDeviceTagResponse {
  code: number;
  message: string;
  data: {
    id: string;
    deleted: boolean;
  };
}

// 获取设备标签列表
export async function getDeviceTagList(params?: {
  page?: number;
  pageSize?: number;
  keyword?: string;
}) {
  return request<DeviceTagListResponse>("/nexlyn/api/tags", {
    method: "GET",
    params,
  });
}

// 创建设备标签
export async function createDeviceTag(data: CreateDeviceTagRequest) {
  return request<CreateDeviceTagResponse>("/nexlyn/api/tags/create", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
  });
}

// 更新设备标签
export async function updateDeviceTag(
  tagId: string,
  data: UpdateDeviceTagRequest
) {
  return request<UpdateDeviceTagResponse>(`/nexlyn/api/tags/${tagId}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    data,
  });
}

// 删除设备标签
export async function deleteDeviceTag(tagId: string) {
  return request<DeleteDeviceTagResponse>(`/nexlyn/api/tags/${tagId}`, {
    method: "DELETE",
  });
}

// GB28181设备信息接口（按实际返回）
export interface GB28181Device {
  deviceId: string;
  name?: string;
  manufacturer?: string;
  model?: string;
  longitude?: string;
  latitude?: string;
  status: "ONLINE" | "OFFLINE" | string;
  mediaIp?: string;
  registerTime?: string;
  updateTime?: string;
  keepAliveTime?: string;
  channelCount: number;
  online?: boolean;
  channels: GB28181Channel[];
  sipIp?: string;
  streamMode?: string;
  password?: string;
  subscribeCatalog?: boolean;
  subscribePosition?: boolean;
  subscribeAlarm?: boolean;
}

// GB28181通道信息接口（按实际返回）
export interface GB28181Channel {
  id?: string;
  deviceId: string;
  channelId: string;
  parentId?: string;
  name: string;
  manufacturer?: string;
  model?: string;
  owner?: string;
  civilCode?: string;
  address?: string;
  port?: number;
  parental?: number;
  safetyWay?: number;
  registerWay?: number;
  secrecy?: number;
  status: "ON" | "OFF" | string;
  gpsTime?: string;
  longitude?: string;
  latitude?: string;
  // 流地址（可能通过其他接口获取，这里预留）
  rtspUrl_Live?: string;
  flvUrl_Live?: string;
  ssrc_Live?: string;
}

// API响应接口（按实际返回）
export interface GB28181ListResponse {
  data: GB28181Device[];
  code: number;
  message: string;
  total?: number;
}

// 设备通道详情响应接口
export interface DeviceChannelsResponse {
  list: GB28181Channel[];
  code: number;
  message: string;
  total?: number;
}

// 基础响应接口
export interface BaseResponse {
  code: number;
  message: string;
}

// ---------------- 媒体管理（流列表/SSE 汇总） ----------------

export interface StreamTrackInfo {
  codec?: string;
  delta?: string;
  meta?: string; // 例如: "fps: 0, resolution: 1920x1088"
  bps?: number;
  bpsOut?: number;
  fps?: number;
  width?: number;
  height?: number;
  gop?: number;
}

export interface StreamItem {
  path: string; // 示例: "deviceId/channelId"
  state: number; // 1/2 等，具体展示转换在页面进行
  subscribers: number;
  audioTrack?: StreamTrackInfo | null;
  videoTrack?: StreamTrackInfo | null;
  startTime?: string;
  pluginName?: string; // 示例: "GB28181"
  type?: string; // live/vod
  meta?: string;
  isPaused?: boolean;
  gop?: number;
  speed?: number;
  bufferTime?: string;
  stopOnIdle?: boolean;
  recording?: Array<string>;
}

export interface StreamListResponse {
  code: number;
  message: string;
  total?: number;
  pageNum?: number;
  pageSize?: number;
  data: StreamItem[];
}

export interface SummarySSEPayload {
  memory?: { total?: number; free?: number; used?: number; usage?: number };
  cpuUsage?: number;
  hardDisk?: { total?: number; free?: number; used?: number; usage?: number };
  netWork?: Array<{
    name?: string;
    receive?: number;
    sent?: number;
    receiveSpeed?: number;
    sentSpeed?: number;
  }>;
  streamCount?: number;
  subscribeCount?: number;
  pullCount?: number;
}

// 获取媒体流列表（来自 mediahandler）
export async function getStreamList(params?: {
  format?: "json";
  page?: number;
  count?: number;
}) {
  return request<StreamListResponse>("/api/stream/list", {
    method: "GET",
    params,
  });
}

// ---------------- 详情与订阅者 ----------------

export interface StreamInfoResponse {
  code: number;
  message: string;
  data: StreamItem;
}

export interface SubscriberReaderInfo {
  sequence?: number;
  timestamp?: number;
  delay?: number;
  state?: number;
  bps?: number;
}

export interface SubscriberItem {
  id: number;
  startTime?: string;
  audioReader?: SubscriberReaderInfo | null;
  videoReader?: SubscriberReaderInfo | null;
  meta?: string;
  bufferTime?: string;
  subMode?: number;
  syncMode?: number;
  pluginName?: string;
  type?: string;
  remoteAddr?: string;
}

export interface SubscribersResponse {
  code: number;
  message: string;
  total?: number;
  pageNum?: number;
  pageSize?: number;
  data: SubscriberItem[];
}

// 获取单个流的详情
export async function getStreamInfo(
  streamPath: string,
  params?: { format?: "json" }
) {
  return request<StreamInfoResponse>(`/api/stream/info/${streamPath}`, {
    method: "GET",
    params,
  });
}

// 获取订阅者列表
export async function getSubscribers(
  streamPath: string,
  params?: { format?: "json"; page?: number; count?: number }
) {
  return request<SubscribersResponse>(`/api/subscribers/${streamPath}`, {
    method: "GET",
    params,
  });
}

// 获取GB28181设备列表
export async function getGB28181DeviceList(params?: {
  page?: number;
  count?: number;
  query?: string;
  status?: boolean;
}) {
  return request<GB28181ListResponse>("/gb28181/api/list", {
    method: "GET",
    params,
  });
}

// 获取设备详情（按新的API路径和响应格式）
export async function getGB28181DeviceDetail(deviceId: string) {
  return request<{
    code: number;
    message: string;
    data: GB28181Device;
  }>(`/gb28181/api/devices/${deviceId}`, {
    method: "GET",
  });
}

// 云台控制（PTZ控制）
export async function ptzControl(
  deviceId: string,
  channelId: string,
  ptzcmd: string
) {
  return request(`/gb28181/api/ptz/${deviceId}/${channelId}`, {
    method: "GET",
    params: { ptzcmd, format: "json" },
  });
}

// 录像控制
export async function recordingControl(
  deviceId: string,
  channelId: string,
  cmdType: "Record" | "RecordStop"
) {
  return request(`/gb28181/api/recording/${cmdType}/${deviceId}/${channelId}`, {
    method: "GET",
  });
}

// 抓拍
export async function snapControl(deviceId: string, channelId: string) {
  return request(`/gb28181/api/snap/${deviceId}/${channelId}`, {
    method: "GET",
  });
}

// 同步设备信息
export async function syncDevice(deviceId: string) {
  return request(`/gb28181/api/devices/${deviceId}/sync`, {
    method: "GET",
  });
}

// 获取设备通道详情
export async function getDeviceChannels(
  deviceId: string,
  params?: {
    page?: number;
    count?: number;
    query?: string;
    online?: boolean;
    channelType?: boolean;
  }
) {
  return request<DeviceChannelsResponse>(
    `/gb28181/api/devices/${deviceId}/channels`,
    {
      method: "GET",
      params,
    }
  );
}

// 删除设备
export async function removeDevice(deviceId: string) {
  return request<BaseResponse>(`/gb28181/api/device/remove/${deviceId}`, {
    method: "DELETE",
  });
}

// ================ Nexlyn API 设备管理接口 ================

// Nexlyn 设备信息接口
export interface NexlynDevice {
  // 设备绑定信息（仅已绑定设备有这些字段）
  id?: string; // 绑定记录主键ID (UUID)
  deviceId: string; // GB28181设备ID
  device_id?: string; // 兼容旧字段名
  tenantId?: string; // 租户ID (UUID)
  organizationId?: string; // 组织ID (UUID)
  deviceAlias?: string; // 设备别名（用户自定义名称）
  device_alias?: string; // 兼容旧字段名
  bindStatus?: "bound" | "unbound"; // 绑定状态
  bindTime?: string; // 设备绑定时间
  bind_time?: string; // 兼容旧字段名
  bindCreatedAt?: string; // 绑定记录创建时间
  bindUpdateTime?: string; // 绑定信息最后更新时间
  tags?: Array<{ id: string; name: string; description?: string }>; // 设备标签数组（与IoT DeviceTagSummary对齐）

  // 设备实时信息
  name?: string; // 设备名称（来自GB28181）
  manufacturer?: string; // 设备制造商
  model?: string; // 设备型号
  status: "ON" | "OFF"; // 设备状态
  online: boolean; // 设备在线状态
  longitude?: string; // 设备经度
  latitude?: string; // 设备纬度
  registerTime?: string; // 设备注册时间
  register_time?: string; // 兼容旧字段名
  deviceUpdateTime?: string; // 设备信息最后更新时间
  update_time?: string; // 兼容旧字段名
  keepaliveTime?: string; // 最后心跳时间
  channelCount?: number; // 通道数量
  channel_count?: number; // 兼容旧字段名
  mediaIp?: string; // 媒体IP地址
  sipIp?: string; // SIP信令IP地址
  password?: string; // 设备密码
  streamMode?: string; // 流模式（UDP/TCP）

  // 其他字段（用于兼容性）
  owner?: string;
  civil_code?: string;
  address?: string;
  port?: number;
  ip?: string;
  channels?: GB28181Channel[];
}

// Nexlyn 设备列表响应
export interface NexlynDeviceListResponse {
  code: number;
  message: string;
  data: {
    total: number;
    pages: number;
    devices: NexlynDevice[];
  };
}

// Nexlyn 设备详情响应
export interface NexlynDeviceDetailResponse {
  code: number;
  message: string;
  data: NexlynDevice;
}

// 设备绑定请求
export interface DeviceBindRequest {
  deviceId: string;
  deviceAlias: string;
  organizationId?: string; // 组织ID
  tags?: string[];
}

// 设备绑定响应
export interface DeviceBindResponse {
  code: number;
  message: string;
  data: {
    deviceId: string;
    success: boolean;
  };
}

// 设备别名更新请求
export interface DeviceAliasUpdateRequest {
  deviceAlias: string;
}

// 设备别名更新响应
export interface DeviceAliasUpdateResponse {
  code: number;
  message: string;
  data: {
    deviceId: string;
    deviceAlias: string;
    updateTime: string;
  };
}

// 设备标签更新请求
export interface DeviceTagsUpdateRequest {
  tags: string[];
}

// 设备标签更新响应
export interface DeviceTagsUpdateResponse {
  code: number;
  message: string;
  data: {
    deviceId: string;
    tags: string[];
    updateTime: string;
  };
}

// 未绑定设备列表响应
export interface UnboundDeviceListResponse {
  code: number;
  message: string;
  data: {
    total: number;
    devices: NexlynDevice[];
  };
}

// 获取Nexlyn设备列表
export async function getNexlynDeviceList(
  params?: {
    page?: number;
    pageSize?: number;
    page_size?: number; // 保持向后兼容
    status?: "ON" | "OFF";
    keyword?: string;
    tags?: string[];
    bindStatus?: "bound" | "unbound"; // 新增绑定状态参数
    organizationId?: string; // 组织ID（单个）
  },
  options?: {
    skipErrorHandler?: boolean;
  }
) {
  return request<NexlynDeviceListResponse>("/nexlyn/api/devices", {
    method: "GET",
    params,
    ...(options || {}),
  });
}

// 获取Nexlyn设备详情
export async function getNexlynDeviceDetail(
  deviceId: string,
  params?: {
    organizationId?: string;
  }
) {
  return request<NexlynDeviceDetailResponse>(
    `/nexlyn/api/devices/${deviceId}`,
    {
      method: "GET",
      params,
    }
  );
}

// 绑定设备
export async function bindDevice(data: DeviceBindRequest) {
  return request<DeviceBindResponse>("/nexlyn/api/devices/bind", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    data,
  });
}

// 编辑设备别名
export async function updateDeviceAlias(
  deviceId: string,
  data: DeviceAliasUpdateRequest
) {
  return request<DeviceAliasUpdateResponse>(
    `/nexlyn/api/devices/${deviceId}/alias`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      data,
    }
  );
}

// 编辑设备标签
export async function updateDeviceTags(
  deviceId: string,
  data: DeviceTagsUpdateRequest
) {
  return request<DeviceTagsUpdateResponse>(
    `/nexlyn/api/devices/${deviceId}/tags`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      data,
    }
  );
}

// 解绑设备
export async function unbindDevice(deviceId: string) {
  return request<BaseResponse>(`/nexlyn/api/devices/${deviceId}/unbind`, {
    method: "DELETE",
  });
}

// 获取未绑定设备列表（内部调用统一接口）
export async function getUnboundDeviceList(params?: {
  page?: number;
  page_size?: number;
  pageSize?: number;
  organizationId?: string;
}) {
  // 使用统一的设备列表接口，设置bindStatus为unbound，并传递组织ID
  return getNexlynDeviceList(
    {
      ...params,
      bindStatus: "unbound",
      organizationId: params?.organizationId,
    },
    { skipErrorHandler: true }
  );
}

// ================ 录像管理相关接口 ================

// 录像记录信息
export interface RecordingItem {
  deviceId: string;
  name: string;
  filePath: string;
  address: string;
  startTime: string;
  endTime: string;
  secrecy: number;
  type: string;
  recorderId: string;
}

// 录像查询请求参数
export interface RecordingQueryParams {
  startTime: number; // 开始时间戳（毫秒）
  endTime: number; // 结束时间戳（毫秒）
}

// 录像查询响应
export interface RecordingQueryResponse {
  code: number;
  message: string;
  data: {
    total: number;
    records: RecordingItem[] | null;
  };
}

// 查询设备通道的录像列表
export async function getRecordings(
  deviceId: string,
  channelId: string,
  params: RecordingQueryParams
) {
  return request<RecordingQueryResponse>(
    `/nexlyn/api/records/${deviceId}/${channelId}`,
    {
      method: "GET",
      params,
    }
  );
}

// ================ 设备统计分析相关接口 ================

// 通道详情信息
export interface ChannelDetail {
  channelId: string; // 通道ID
  name: string; // 通道名称
  status: "ON" | "OFF" | "DEVICE_OFFLINE"; // 通道状态
}

// 设备详情信息
export interface DeviceDetail {
  deviceId: string; // 设备ID
  deviceName: string; // 设备名称
  online: boolean; // 设备在线状态
  channelCount: number; // 通道总数
  channels: ChannelDetail[]; // 通道列表
}

// 设备统计分组信息
export interface DeviceStatisticsGroup {
  groupKey: string; // 分组键（组织ID或标签名）
  groupName: string; // 分组名称
  deviceTotal: number; // 总设备数
  deviceOnline: number; // 在线设备数
  deviceOffline: number; // 离线设备数
  channelTotal?: number; // 总通道数
  channelOnline?: number; // 在线通道数
  channelOffline?: number; // 离线通道数
}

// 设备统计数据
export interface DeviceStatisticsData {
  deviceTotal: number; // 总设备数
  deviceOnline: number; // 在线设备数
  deviceOffline: number; // 离线设备数
  channelTotal?: number; // 总通道数
  channelOnline?: number; // 在线通道数
  channelOffline?: number; // 离线通道数
  devices?: DeviceDetail[]; // 设备详情列表（可选）
  groups?: DeviceStatisticsGroup[]; // 分组统计（可选）
}

// 设备统计响应
export interface DeviceStatisticsResponse {
  code: number;
  message: string;
  data: DeviceStatisticsData;
}

// 设备统计请求参数
export interface DeviceStatisticsParams {
  organizationId: string; // 组织ID（必填）
  deviceIds?: string[]; // 设备ID列表（可选，多个用逗号分隔）
  groupBy?: "org" | "tag"; // 分组方式：org=按组织分组，tag=按标签分组
}

// 获取设备统计分析（包含设备和通道统计）
export async function getDeviceStatistics(params: DeviceStatisticsParams) {
  return request<DeviceStatisticsResponse>("/api/v1/video/devices/statistics", {
    method: "GET",
    params: {
      ...params,
      deviceIds: params.deviceIds?.join(","), // 将数组转换为逗号分隔的字符串
    },
  });
}
