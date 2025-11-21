/**
 * 设备绑定相关类型
 */

// 设备标签摘要（用于设备列表展示）
export interface DeviceTagSummary {
  id: string; // 标签ID
  name: string; // 标签名称
  color: string; // 标签颜色
  description: string; // 标签描述
}

// 设备绑定完整信息
export interface DeviceBinding {
  id: string; // 设备绑定记录ID(UUID)
  deviceId: string; // 设备唯一标识符
  deviceName: string; // 设备名称
  deviceAlias: string; // 设备别名
  deviceModel: string; // 设备型号(关联sensor_templates.model)
  deviceCategory: string; // 设备类别(冗余字段)
  description: string; // 设备描述
  location: string; // 安装位置
  installationDate: string; // 安装日期(YYYY-MM-DD)
  status: string; // 设备状态(active/inactive/maintenance/error/decommissioned)
  isOnline: boolean; // 当前在线状态(从Redis实时查询)
  lastDataAt: string; // 最后接收数据时间
  tenantId: number; // 所属租户ID
  orgId: string; // 所属组织ID(UUID)
  orgName?: string; // 所属组织名称
  tags: DeviceTagSummary[]; // 关联的标签列表
  createdBy: number; // 创建者用户ID
  updatedBy: number; // 最后修改者用户ID
  createdAt: string; // 创建时间
  updatedAt: string; // 更新时间
}

// 设备绑定摘要信息(用于列表展示)
export interface DeviceBindingSummary {
  id: string; // 设备绑定记录ID
  deviceId: string; // 设备唯一标识符
  deviceName: string; // 设备名称
  deviceAlias: string; // 设备别名
  deviceModel: string; // 设备型号
  deviceCategory: string; // 设备类别
  location: string; // 安装位置
  installationDate?: string; // 安装日期(YYYY-MM-DD)
  status: string; // 设备状态
  isOnline: boolean; // 当前在线状态(从Redis实时查询)
  lastDataAt: string; // 最后接收数据时间
  orgId: string; // 所属组织ID
  orgName?: string; // 所属组织名称
  tags: DeviceTagSummary[]; // 关联的标签列表
  createdAt: string; // 创建时间
}

// 未绑定设备信息
export interface UnboundDevice {
  deviceId: string; // 设备ID
  category: string; // 设备类别
  isOnline: boolean; // 在线状态
  lastSeen: string; // 最后上报时间
}
