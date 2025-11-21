/**
 * 设备标签相关类型
 */

// 设备标签完整信息
export interface DeviceTag {
  id: string; // 标签ID(UUID)
  name: string; // 标签名称
  color: string; // 标签颜色(HEX格式,如:#FF5733)
  icon: string; // 标签图标
  description: string; // 标签描述
  tenantId: number; // 所属租户ID
  deviceCount: number; // 使用该标签的设备数量
  createdBy: number; // 创建者用户ID
  updatedBy: number; // 最后修改者用户ID
  createdAt: string; // 创建时间
  updatedAt: string; // 更新时间
}
