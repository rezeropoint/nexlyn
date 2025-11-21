/**
 * 设备类别映射工具
 * 提供设备类别的显示名称、图标等信息
 */

export interface DeviceCategoryInfo {
  code: string;
  name: string;
  icon: string;
  color: string;
}

// 设备类别映射表
export const DEVICE_CATEGORY_MAP: Record<string, DeviceCategoryInfo> = {
  // 摄像头
  camera: {
    code: "camera",
    name: "监控设备",
    icon: "VideoCameraOutlined",
    color: "#1890ff",
  },

  // IoT传感器类别
  temperature_sensor: {
    code: "temperature_sensor",
    name: "温度传感器",
    icon: "DashboardOutlined",
    color: "#ff4d4f",
  },
  humidity_sensor: {
    code: "humidity_sensor",
    name: "湿度传感器",
    icon: "CloudOutlined",
    color: "#13c2c2",
  },
  pressure_sensor: {
    code: "pressure_sensor",
    name: "压力传感器",
    icon: "DashboardOutlined",
    color: "#722ed1",
  },
  light_sensor: {
    code: "light_sensor",
    name: "光照传感器",
    icon: "BulbOutlined",
    color: "#faad14",
  },
  motion_sensor: {
    code: "motion_sensor",
    name: "运动传感器",
    icon: "RadarChartOutlined",
    color: "#52c41a",
  },
  smoke_detector: {
    code: "smoke_detector",
    name: "烟雾探测器",
    icon: "FireOutlined",
    color: "#f5222d",
  },
  door_sensor: {
    code: "door_sensor",
    name: "门磁传感器",
    icon: "ExportOutlined",
    color: "#2f54eb",
  },
  water_leak_sensor: {
    code: "water_leak_sensor",
    name: "漏水传感器",
    icon: "AlertOutlined",
    color: "#1890ff",
  },
  air_quality_sensor: {
    code: "air_quality_sensor",
    name: "空气质量传感器",
    icon: "ExperimentOutlined",
    color: "#52c41a",
  },
  ai_box: {
    code: "ai_box",
    name: "AI盒子",
    icon: "RobotOutlined",
    color: "#722ed1",
  },
  gateway: {
    code: "gateway",
    name: "网关",
    icon: "ClusterOutlined",
    color: "#13c2c2",
  },

  // 未知类别
  unknown: {
    code: "unknown",
    name: "未知设备",
    icon: "QuestionCircleOutlined",
    color: "#8c8c8c",
  },
};

/**
 * 获取设备类别信息
 * @param categoryCode 设备类别代码
 * @returns 设备类别信息
 */
export function getDeviceCategoryInfo(
  categoryCode: string
): DeviceCategoryInfo {
  return DEVICE_CATEGORY_MAP[categoryCode] || DEVICE_CATEGORY_MAP.unknown;
}

/**
 * 获取所有IoT设备类别（不包括摄像头）
 * @returns IoT设备类别列表
 */
export function getIoTCategories(): DeviceCategoryInfo[] {
  return Object.values(DEVICE_CATEGORY_MAP).filter(
    (category) => category.code !== "camera" && category.code !== "unknown"
  );
}

/**
 * 判断是否为IoT设备类别
 * @param categoryCode 设备类别代码
 * @returns 是否为IoT设备
 */
export function isIoTCategory(categoryCode: string): boolean {
  return categoryCode !== "camera" && categoryCode !== "unknown";
}
