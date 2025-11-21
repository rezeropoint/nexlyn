// 视频管理相关类型定义

declare namespace API {
  // ================ 设备标签管理 ================

  /** 设备标签信息 */
  type DeviceTag = {
    id: string;
    tagName: string;
    description?: string;
    deviceCount: number;
    createdAt: string;
    updatedAt: string;
  };

  /** 标签列表响应 */
  type DeviceTagListResponse = {
    code: number;
    message: string;
    data: {
      total: number;
      pages: number;
      tags: DeviceTag[];
    };
  };

  /** 创建标签请求 */
  type CreateDeviceTagRequest = {
    tagName: string;
    description?: string;
  };

  /** 更新标签请求 */
  type UpdateDeviceTagRequest = {
    tagName?: string;
    description?: string;
  };

  // ================ GB28181设备管理 ================

  /** GB28181设备信息 */
  type GB28181Device = {
    device_id: string;
    device_name: string;
    manufacturer: string;
    model: string;
    firmware: string;
    transport: string;
    status: "ON" | "OFF";
    created_at: string;
    updated_at: string;
    last_keepalive_at?: string;
    channels_count?: {
      total: number;
      online: number;
      offline: number;
    };
  };

  /** 设备列表响应 */
  type GB28181DeviceListResponse = BaseResponse & {
    data: {
      total: number;
      pages: number;
      devices: GB28181Device[];
    };
  };

  /** 设备详情响应 */
  type GB28181DeviceDetailResponse = BaseResponse & {
    data: GB28181Device;
  };

  /** 设备通道信息 */
  type DeviceChannel = {
    device_id: string;
    channel_id: string;
    channel_name: string;
    manufacturer: string;
    model: string;
    owner: string;
    civil_code: string;
    block?: string;
    address?: string;
    parental: number;
    parent_id?: string;
    safety_way: number;
    register_way: number;
    cert_num?: string;
    certifiable: number;
    err_code?: number;
    end_time?: string;
    secrecy: number;
    ip_address?: string;
    port?: number;
    password?: string;
    status: "ON" | "OFF";
    longitude?: number;
    latitude?: number;
    created_at: string;
    updated_at: string;
  };

  /** 设备通道列表响应 */
  type DeviceChannelsResponse = BaseResponse & {
    list: DeviceChannel[];
    total: number;
  };

  /** 设备同步响应 */
  type SyncDeviceResponse = BaseResponse;

  /** 设备删除响应 */
  type RemoveDeviceResponse = BaseResponse;

  // ================ 媒体流管理 ================

  /** 媒体流信息 */
  type MediaStream = {
    app: string;
    stream: string;
    device_id?: string;
    channel_id?: string;
    ssrc?: string;
    rtp_ip?: string;
    rtp_port?: number;
    tcp_mode?: boolean;
    created_at: string;
    updated_at: string;
    viewer_count?: number;
  };

  /** 媒体流列表响应 */
  type MediaStreamListResponse = BaseResponse & {
    data: {
      total: number;
      streams: MediaStream[];
    };
  };

  /** 播放地址信息 */
  type PlayUrl = {
    flv?: string;
    hls?: string;
    rtmp?: string;
    webrtc?: string;
  };

  /** 获取播放地址响应 */
  type GetPlayUrlResponse = BaseResponse & {
    data: PlayUrl;
  };

  /** 推流请求 */
  type StartStreamRequest = {
    device_id: string;
    channel_id: string;
  };

  /** 停流请求 */
  type StopStreamRequest = {
    device_id: string;
    channel_id: string;
  };

  // ================ 设备绑定管理 ================

  /** 设备绑定请求 */
  type BindDeviceRequest = {
    device_id: string;
    tag_ids?: string[];
  };

  /** 设备别名更新请求 */
  type UpdateDeviceAliasRequest = {
    device_alias: string;
  };

  /** 设备标签更新请求 */
  type UpdateDeviceTagsRequest = {
    tag_ids: string[];
  };
}
