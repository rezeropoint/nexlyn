/**
 * 逻辑图图标配置
 * 使用 react-icons 库，支持多个图标集
 */

import {
  FaBell,
  FaDatabase,
  FaLightbulb,
  FaRobot,
  FaShieldAlt,
  FaSitemap,
  FaVideo,
} from 'react-icons/fa';
import {
  MdAnalytics,
  MdDashboard,
  MdNotifications,
  MdSecurity,
  MdSensors,
  MdSettings,
} from 'react-icons/md';
import {
  BiData,
  BiGitBranch,
  BiNetworkChart,
  BiTransfer,
} from 'react-icons/bi';
import {
  AiOutlineCloudServer,
  AiOutlineControl,
  AiOutlineProject,
  AiOutlineSchedule,
} from 'react-icons/ai';
import {
  GiArtificialIntelligence,
  GiElectric,
  GiProcessor,
} from 'react-icons/gi';
import {
  IoMdCube,
  IoMdGitNetwork,
} from 'react-icons/io';
import { RiFlowChart } from 'react-icons/ri';
import { TbBrain, TbCircuitSwitchClosed } from 'react-icons/tb';
import React from 'react';
import type { GlobalToken } from 'antd/es/theme/interface';

/**
 * 颜色语义 Key
 */
export type ColorKey = 'primary' | 'success' | 'warning' | 'error' | 'purple' | 'magenta' | 'cyan' | 'geekblue' | 'textSecondary';

/**
 * 根据 token 获取颜色值
 */
export const getColorByKey = (colorKey: ColorKey, token: GlobalToken): string => {
  const map: Record<ColorKey, string> = {
    primary: token.colorPrimary,
    success: token.colorSuccess,
    warning: token.colorWarning,
    error: token.colorError,
    purple: token.purple,
    magenta: token.magenta,
    cyan: token.cyan,
    geekblue: token.geekblue,
    textSecondary: token.colorTextSecondary,
  };
  return map[colorKey];
};

/**
 * 图标配置项
 */
export interface IconOption {
  name: string; // 图标名称（唯一标识）
  label: string; // 显示名称
  component: React.ComponentType<any>; // 图标组件
  category: string; // 图标分类
  colorKey?: ColorKey; // 颜色语义 key
}

/**
 * 图标分类
 */
export enum IconCategory {
  AI = 'AI/智能',
  IoT = '物联网',
  Data = '数据处理',
  Workflow = '流程/工作流',
  Monitor = '监控/告警',
  Security = '安全',
  Integration = '集成/连接',
  Media = '媒体',
  Other = '其他',
}

/**
 * 预定义图标库
 */
export const ICON_OPTIONS: IconOption[] = [
  // AI/智能类
  {
    name: 'FaRobot',
    label: '机器人',
    component: FaRobot,
    category: IconCategory.AI,
    colorKey: 'primary',
  },
  {
    name: 'GiArtificialIntelligence',
    label: 'AI智能',
    component: GiArtificialIntelligence,
    category: IconCategory.AI,
    colorKey: 'purple',
  },
  {
    name: 'TbBrain',
    label: '智能大脑',
    component: TbBrain,
    category: IconCategory.AI,
    colorKey: 'magenta',
  },
  {
    name: 'FaLightbulb',
    label: '智能灯泡',
    component: FaLightbulb,
    category: IconCategory.AI,
    colorKey: 'warning',
  },

  // 物联网类
  {
    name: 'MdSensors',
    label: '传感器',
    component: MdSensors,
    category: IconCategory.IoT,
    colorKey: 'cyan',
  },
  {
    name: 'IoMdCube',
    label: '设备',
    component: IoMdCube,
    category: IconCategory.IoT,
    colorKey: 'geekblue',
  },
  {
    name: 'GiElectric',
    label: '电气',
    component: GiElectric,
    category: IconCategory.IoT,
    colorKey: 'warning',
  },
  {
    name: 'TbCircuitSwitchClosed',
    label: '电路',
    component: TbCircuitSwitchClosed,
    category: IconCategory.IoT,
    colorKey: 'success',
  },

  // 数据处理类
  {
    name: 'FaDatabase',
    label: '数据库',
    component: FaDatabase,
    category: IconCategory.Data,
    colorKey: 'primary',
  },
  {
    name: 'BiData',
    label: '数据',
    component: BiData,
    category: IconCategory.Data,
    colorKey: 'cyan',
  },
  {
    name: 'MdAnalytics',
    label: '分析',
    component: MdAnalytics,
    category: IconCategory.Data,
    colorKey: 'success',
  },
  {
    name: 'GiProcessor',
    label: '处理器',
    component: GiProcessor,
    category: IconCategory.Data,
    colorKey: 'purple',
  },

  // 流程/工作流类
  {
    name: 'RiFlowChart',
    label: '流程图',
    component: RiFlowChart,
    category: IconCategory.Workflow,
    colorKey: 'primary',
  },
  {
    name: 'AiOutlineProject',
    label: '项目',
    component: AiOutlineProject,
    category: IconCategory.Workflow,
    colorKey: 'success',
  },
  {
    name: 'BiGitBranch',
    label: '分支',
    component: BiGitBranch,
    category: IconCategory.Workflow,
    colorKey: 'warning',
  },
  {
    name: 'AiOutlineSchedule',
    label: '调度',
    component: AiOutlineSchedule,
    category: IconCategory.Workflow,
    colorKey: 'magenta',
  },
  {
    name: 'FaSitemap',
    label: '拓扑',
    component: FaSitemap,
    category: IconCategory.Workflow,
    colorKey: 'purple',
  },

  // 监控/告警类
  {
    name: 'MdDashboard',
    label: '仪表盘',
    component: MdDashboard,
    category: IconCategory.Monitor,
    colorKey: 'primary',
  },
  {
    name: 'MdNotifications',
    label: '通知',
    component: MdNotifications,
    category: IconCategory.Monitor,
    colorKey: 'warning',
  },
  {
    name: 'FaBell',
    label: '告警',
    component: FaBell,
    category: IconCategory.Monitor,
    colorKey: 'error',
  },
  {
    name: 'AiOutlineControl',
    label: '控制',
    component: AiOutlineControl,
    category: IconCategory.Monitor,
    colorKey: 'cyan',
  },

  // 安全类
  {
    name: 'MdSecurity',
    label: '安全',
    component: MdSecurity,
    category: IconCategory.Security,
    colorKey: 'success',
  },
  {
    name: 'FaShieldAlt',
    label: '防护',
    component: FaShieldAlt,
    category: IconCategory.Security,
    colorKey: 'primary',
  },

  // 集成/连接类
  {
    name: 'BiNetworkChart',
    label: '网络',
    component: BiNetworkChart,
    category: IconCategory.Integration,
    colorKey: 'primary',
  },
  {
    name: 'IoMdGitNetwork',
    label: '节点网络',
    component: IoMdGitNetwork,
    category: IconCategory.Integration,
    colorKey: 'purple',
  },
  {
    name: 'BiTransfer',
    label: '数据传输',
    component: BiTransfer,
    category: IconCategory.Integration,
    colorKey: 'cyan',
  },
  {
    name: 'AiOutlineCloudServer',
    label: '云服务',
    component: AiOutlineCloudServer,
    category: IconCategory.Integration,
    colorKey: 'geekblue',
  },

  // 媒体类
  {
    name: 'FaVideo',
    label: '视频',
    component: FaVideo,
    category: IconCategory.Media,
    colorKey: 'magenta',
  },

  // 其他
  {
    name: 'MdSettings',
    label: '配置',
    component: MdSettings,
    category: IconCategory.Other,
    colorKey: 'textSecondary',
  },
];

/**
 * 根据图标名称获取图标组件
 */
export const getIconComponent = (iconName?: string): React.ComponentType<any> | null => {
  if (!iconName) return null;
  const iconOption = ICON_OPTIONS.find((option) => option.name === iconName);
  return iconOption?.component || null;
};

/**
 * 根据图标名称获取颜色 key
 */
export const getIconColorKey = (iconName?: string): ColorKey | undefined => {
  if (!iconName) return undefined;
  const iconOption = ICON_OPTIONS.find((option) => option.name === iconName);
  return iconOption?.colorKey;
};

/**
 * 根据图标名称和 token 获取实际颜色值
 */
export const getIconColor = (iconName: string | undefined, token: GlobalToken): string | undefined => {
  const colorKey = getIconColorKey(iconName);
  if (!colorKey) return undefined;
  return getColorByKey(colorKey, token);
};

/**
 * 按分类分组图标
 */
export const getIconsByCategory = (): Record<string, IconOption[]> => {
  return ICON_OPTIONS.reduce((acc, icon) => {
    if (!acc[icon.category]) {
      acc[icon.category] = [];
    }
    acc[icon.category].push(icon);
    return acc;
  }, {} as Record<string, IconOption[]>);
};
