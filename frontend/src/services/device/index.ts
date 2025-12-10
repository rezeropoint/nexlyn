import { listDevices, type ListDevicesRequest } from "@/services/iot";
import {
  getDeviceStatistics as getVideoDeviceStatistics,
  type ChannelDetail,
  type DeviceDetail,
  type DeviceStatisticsGroup,
  type DeviceStatisticsData as VideoDeviceStatisticsData,
  type DeviceStatisticsParams as VideoDeviceStatisticsParams,
} from "@/services/video";

// ===== 统一设备类型定义 =====

export type DeviceType = "camera" | "iot_sensor" | "all";

export interface DeviceCategory {
  type: DeviceType;
  name: string;
  icon: string;
}

// 设备类别配置
export const DEVICE_CATEGORIES: DeviceCategory[] = [
  { type: "all", name: "全部设备", icon: "AppstoreOutlined" },
  { type: "camera", name: "监控设备", icon: "VideoCameraOutlined" },
  { type: "iot_sensor", name: "IoT传感器", icon: "ApiOutlined" },
];

// ===== 统一设备统计数据结构 =====

export interface DeviceCategoryStats {
  total: number;
  online: number;
  offline: number;
}

export interface UnifiedDeviceStatisticsData {
  deviceTotal: number; // 总设备数
  deviceOnline: number; // 在线设备数
  deviceOffline: number; // 离线设备数
  channelTotal?: number; // 总通道数
  channelOnline?: number; // 在线通道数
  channelOffline?: number; // 离线通道数
  devices?: DeviceDetail[]; // 设备详情列表（可选）
  groups?: DeviceStatisticsGroup[]; // 分组统计（可选）
  byType?: {
    camera: DeviceCategoryStats;
    iotSensor: DeviceCategoryStats;
  };
  byCategory?: Record<string, DeviceCategoryStats>; // 按设备类别细分（包括摄像头和各种IoT设备类别）
}

export interface UnifiedDeviceStatisticsParams {
  organizationId: string; // 组织ID（必填）
  deviceType?: DeviceType; // 设备类型过滤
  deviceIds?: string[]; // 设备ID列表（可选）
  groupBy?: "org" | "tag" | "type"; // 分组方式：org=按组织分组，tag=按标签分组，type=按设备类型分组
}

// ===== 统一设备统计API =====

/**
 * 获取摄像头设备统计
 * 注意：目前后端未实现 /api/v1/video/devices/statistics 接口，暂时返回空数据
 */
async function getCameraStatistics(
  _params: VideoDeviceStatisticsParams
): Promise<VideoDeviceStatisticsData> {
  // TODO: 等后端实现视频设备统计接口后再启用
  // try {
  //   const response = await getVideoDeviceStatistics(params);
  //   if (response.code === 0 && response.data) {
  //     return response.data;
  //   }
  //   return { deviceTotal: 0, deviceOnline: 0, deviceOffline: 0 };
  // } catch (error) {
  //   console.error("获取摄像头统计失败:", error);
  //   return { deviceTotal: 0, deviceOnline: 0, deviceOffline: 0 };
  // }
  return { deviceTotal: 0, deviceOnline: 0, deviceOffline: 0 };
}

/**
 * 获取IoT传感器设备统计
 */
async function getIoTSensorStatistics(params: {
  organizationId: string;
  groupBy?: "org" | "tag";
}): Promise<{
  deviceTotal: number;
  deviceOnline: number;
  deviceOffline: number;
  groups?: DeviceStatisticsGroup[];
  byCategory?: Record<string, DeviceCategoryStats>; // 按设备类别统计
}> {
  try {
    const requestParams: ListDevicesRequest = {
      orgId: params.organizationId,
      page: 1,
      pageSize: 10000, // 获取全部设备用于统计
    };

    const response = await listDevices(requestParams);

    if (response.code === 0 && response.data?.list) {
      const devices = response.data.list;
      const deviceTotal = devices.length;
      const deviceOnline = devices.filter((d) => d.isOnline).length;
      const deviceOffline = deviceTotal - deviceOnline;

      // 按设备类别统计
      const categoryStats: Record<string, DeviceCategoryStats> = {};
      devices.forEach((device) => {
        const category = device.deviceCategory || "unknown";
        if (!categoryStats[category]) {
          categoryStats[category] = { total: 0, online: 0, offline: 0 };
        }
        categoryStats[category].total++;
        if (device.isOnline) {
          categoryStats[category].online++;
        } else {
          categoryStats[category].offline++;
        }
      });

      // 如果需要分组统计
      let groups: DeviceStatisticsGroup[] | undefined;
      if (params.groupBy === "org") {
        // 按组织分组
        const orgMap = new Map<
          string,
          { name: string; total: number; online: number }
        >();
        devices.forEach((device) => {
          const orgId = device.orgId;
          if (!orgMap.has(orgId)) {
            // 优先使用orgName，如果没有则使用orgId
            orgMap.set(orgId, {
              name: device.orgName || orgId,
              total: 0,
              online: 0,
            });
          }
          const stat = orgMap.get(orgId)!;
          stat.total++;
          if (device.isOnline) stat.online++;
        });

        groups = Array.from(orgMap.entries()).map(([orgId, stat]) => ({
          groupKey: orgId,
          groupName: stat.name,
          deviceTotal: stat.total,
          deviceOnline: stat.online,
          deviceOffline: stat.total - stat.online,
        }));
      } else if (params.groupBy === "tag") {
        // 按标签分组
        const tagMap = new Map<
          string,
          { name: string; total: number; online: number }
        >();
        devices.forEach((device) => {
          if (device.tags && device.tags.length > 0) {
            device.tags.forEach((tag) => {
              if (!tagMap.has(tag.id)) {
                tagMap.set(tag.id, { name: tag.name, total: 0, online: 0 });
              }
              const stat = tagMap.get(tag.id)!;
              stat.total++;
              if (device.isOnline) stat.online++;
            });
          } else {
            // 无标签设备
            if (!tagMap.has("__no_tag__")) {
              tagMap.set("__no_tag__", { name: "未分类", total: 0, online: 0 });
            }
            const stat = tagMap.get("__no_tag__")!;
            stat.total++;
            if (device.isOnline) stat.online++;
          }
        });

        groups = Array.from(tagMap.entries()).map(([tagId, stat]) => ({
          groupKey: tagId,
          groupName: stat.name,
          deviceTotal: stat.total,
          deviceOnline: stat.online,
          deviceOffline: stat.total - stat.online,
        }));
      }

      return {
        deviceTotal,
        deviceOnline,
        deviceOffline,
        groups,
        byCategory: categoryStats,
      };
    }

    return {
      deviceTotal: 0,
      deviceOnline: 0,
      deviceOffline: 0,
      byCategory: {},
    };
  } catch (error) {
    console.error("获取IoT传感器统计失败:", error);
    return {
      deviceTotal: 0,
      deviceOnline: 0,
      deviceOffline: 0,
      byCategory: {},
    };
  }
}

/**
 * 获取统一的设备统计（合并摄像头和IoT传感器）
 * 注意：摄像头统计已包含通道数据
 */
export async function getUnifiedDeviceStatistics(
  params: UnifiedDeviceStatisticsParams
): Promise<UnifiedDeviceStatisticsData> {
  const { organizationId, deviceType = "all", groupBy } = params;

  try {
    // 如果按类型分组，不传递groupBy给后端API（type分组在前端处理）
    const apiGroupBy = groupBy === "type" ? undefined : groupBy;

    // 根据设备类型获取统计数据
    if (deviceType === "iot_sensor") {
      // 仅IoT传感器（没有通道数据）
      const iotStats = await getIoTSensorStatistics({
        organizationId,
        groupBy: apiGroupBy,
      });
      return {
        ...iotStats,
        channelTotal: 0,
        channelOnline: 0,
        channelOffline: 0,
        byType: {
          camera: { total: 0, online: 0, offline: 0 },
          iotSensor: {
            total: iotStats.deviceTotal,
            online: iotStats.deviceOnline,
            offline: iotStats.deviceOffline,
          },
        },
      };
    } else if (deviceType === "camera") {
      // 仅摄像头（包含通道统计）
      const cameraStats = await getCameraStatistics({
        organizationId,
        groupBy: apiGroupBy,
      });
      return {
        deviceTotal: cameraStats.deviceTotal,
        deviceOnline: cameraStats.deviceOnline,
        deviceOffline: cameraStats.deviceOffline,
        channelTotal: cameraStats.channelTotal || 0,
        channelOnline: cameraStats.channelOnline || 0,
        channelOffline: cameraStats.channelOffline || 0,
        devices: cameraStats.devices,
        groups: cameraStats.groups,
        byType: {
          camera: {
            total: cameraStats.deviceTotal,
            online: cameraStats.deviceOnline,
            offline: cameraStats.deviceOffline,
          },
          iotSensor: { total: 0, online: 0, offline: 0 },
        },
        byCategory: {
          camera: {
            total: cameraStats.deviceTotal,
            online: cameraStats.deviceOnline,
            offline: cameraStats.deviceOffline,
          },
        },
      };
    } else {
      // 全部设备：并行请求摄像头（含通道）和IoT传感器统计
      const [cameraStats, iotStats] = await Promise.all([
        getCameraStatistics({ organizationId, groupBy: apiGroupBy }),
        getIoTSensorStatistics({ organizationId, groupBy: apiGroupBy }),
      ]);

      // 合并统计数据
      const mergedData: UnifiedDeviceStatisticsData = {
        deviceTotal: cameraStats.deviceTotal + iotStats.deviceTotal,
        deviceOnline: cameraStats.deviceOnline + iotStats.deviceOnline,
        deviceOffline: cameraStats.deviceOffline + iotStats.deviceOffline,
        // 通道数据来自摄像头统计
        channelTotal: cameraStats.channelTotal || 0,
        channelOnline: cameraStats.channelOnline || 0,
        channelOffline: cameraStats.channelOffline || 0,
        devices: cameraStats.devices || [],
        byType: {
          camera: {
            total: cameraStats.deviceTotal,
            online: cameraStats.deviceOnline,
            offline: cameraStats.deviceOffline,
          },
          iotSensor: {
            total: iotStats.deviceTotal,
            online: iotStats.deviceOnline,
            offline: iotStats.deviceOffline,
          },
        },
        // 按设备类别细分（包括摄像头和所有IoT设备类别）
        byCategory: {
          camera: {
            total: cameraStats.deviceTotal,
            online: cameraStats.deviceOnline,
            offline: cameraStats.deviceOffline,
          },
          ...(iotStats.byCategory || {}),
        },
      };

      // 合并分组数据
      if (groupBy === "type") {
        // 按设备类型分组
        mergedData.groups = [
          {
            groupKey: "camera",
            groupName: "监控设备",
            deviceTotal: cameraStats.deviceTotal,
            deviceOnline: cameraStats.deviceOnline,
            deviceOffline: cameraStats.deviceOffline,
            channelTotal: cameraStats.channelTotal || 0,
            channelOnline: cameraStats.channelOnline || 0,
            channelOffline: cameraStats.channelOffline || 0,
          },
          {
            groupKey: "iot_sensor",
            groupName: "IoT传感器",
            deviceTotal: iotStats.deviceTotal,
            deviceOnline: iotStats.deviceOnline,
            deviceOffline: iotStats.deviceOffline,
            channelTotal: 0,
            channelOnline: 0,
            channelOffline: 0,
          },
        ];
      } else if (groupBy === "org" || groupBy === "tag") {
        // 按组织或标签分组：合并两个来源的分组数据
        const groupMap = new Map<string, DeviceStatisticsGroup>();

        // 添加摄像头分组（包含通道数据）
        if (cameraStats.groups) {
          cameraStats.groups.forEach((group) => {
            groupMap.set(group.groupKey, { ...group });
          });
        }

        // 合并IoT传感器分组（设备数据叠加，通道数据保持不变）
        if (iotStats.groups) {
          iotStats.groups.forEach((group) => {
            if (groupMap.has(group.groupKey)) {
              const existing = groupMap.get(group.groupKey)!;
              existing.deviceTotal += group.deviceTotal;
              existing.deviceOnline += group.deviceOnline;
              existing.deviceOffline += group.deviceOffline;
              // 通道数据已经在摄像头分组中，不需要修改
            } else {
              groupMap.set(group.groupKey, {
                ...group,
                channelTotal: 0,
                channelOnline: 0,
                channelOffline: 0,
              });
            }
          });
        }

        mergedData.groups = Array.from(groupMap.values());
      }

      return mergedData;
    }
  } catch (error) {
    console.error("获取统一设备统计失败:", error);
    return {
      deviceTotal: 0,
      deviceOnline: 0,
      deviceOffline: 0,
      channelTotal: 0,
      channelOnline: 0,
      channelOffline: 0,
      byType: {
        camera: { total: 0, online: 0, offline: 0 },
        iotSensor: { total: 0, online: 0, offline: 0 },
      },
    };
  }
}

// 导出原有类型以保持兼容性
export type { ChannelDetail, DeviceDetail, DeviceStatisticsGroup };
