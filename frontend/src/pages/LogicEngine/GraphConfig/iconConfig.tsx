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

/**
 * 图标配置项
 */
export interface IconOption {
  name: string; // 图标名称（唯一标识）
  label: string; // 显示名称
  component: React.ComponentType<any>; // 图标组件
  category: string; // 图标分类
  color?: string; // 默认颜色
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
    color: '#1890ff',
  },
  {
    name: 'GiArtificialIntelligence',
    label: 'AI智能',
    component: GiArtificialIntelligence,
    category: IconCategory.AI,
    color: '#722ed1',
  },
  {
    name: 'TbBrain',
    label: '智能大脑',
    component: TbBrain,
    category: IconCategory.AI,
    color: '#eb2f96',
  },
  {
    name: 'FaLightbulb',
    label: '智能灯泡',
    component: FaLightbulb,
    category: IconCategory.AI,
    color: '#faad14',
  },

  // 物联网类
  {
    name: 'MdSensors',
    label: '传感器',
    component: MdSensors,
    category: IconCategory.IoT,
    color: '#13c2c2',
  },
  {
    name: 'IoMdCube',
    label: '设备',
    component: IoMdCube,
    category: IconCategory.IoT,
    color: '#2f54eb',
  },
  {
    name: 'GiElectric',
    label: '电气',
    component: GiElectric,
    category: IconCategory.IoT,
    color: '#faad14',
  },
  {
    name: 'TbCircuitSwitchClosed',
    label: '电路',
    component: TbCircuitSwitchClosed,
    category: IconCategory.IoT,
    color: '#52c41a',
  },

  // 数据处理类
  {
    name: 'FaDatabase',
    label: '数据库',
    component: FaDatabase,
    category: IconCategory.Data,
    color: '#1890ff',
  },
  {
    name: 'BiData',
    label: '数据',
    component: BiData,
    category: IconCategory.Data,
    color: '#13c2c2',
  },
  {
    name: 'MdAnalytics',
    label: '分析',
    component: MdAnalytics,
    category: IconCategory.Data,
    color: '#52c41a',
  },
  {
    name: 'GiProcessor',
    label: '处理器',
    component: GiProcessor,
    category: IconCategory.Data,
    color: '#722ed1',
  },

  // 流程/工作流类
  {
    name: 'RiFlowChart',
    label: '流程图',
    component: RiFlowChart,
    category: IconCategory.Workflow,
    color: '#1890ff',
  },
  {
    name: 'AiOutlineProject',
    label: '项目',
    component: AiOutlineProject,
    category: IconCategory.Workflow,
    color: '#52c41a',
  },
  {
    name: 'BiGitBranch',
    label: '分支',
    component: BiGitBranch,
    category: IconCategory.Workflow,
    color: '#fa8c16',
  },
  {
    name: 'AiOutlineSchedule',
    label: '调度',
    component: AiOutlineSchedule,
    category: IconCategory.Workflow,
    color: '#eb2f96',
  },
  {
    name: 'FaSitemap',
    label: '拓扑',
    component: FaSitemap,
    category: IconCategory.Workflow,
    color: '#722ed1',
  },

  // 监控/告警类
  {
    name: 'MdDashboard',
    label: '仪表盘',
    component: MdDashboard,
    category: IconCategory.Monitor,
    color: '#1890ff',
  },
  {
    name: 'MdNotifications',
    label: '通知',
    component: MdNotifications,
    category: IconCategory.Monitor,
    color: '#fa8c16',
  },
  {
    name: 'FaBell',
    label: '告警',
    component: FaBell,
    category: IconCategory.Monitor,
    color: '#f5222d',
  },
  {
    name: 'AiOutlineControl',
    label: '控制',
    component: AiOutlineControl,
    category: IconCategory.Monitor,
    color: '#13c2c2',
  },

  // 安全类
  {
    name: 'MdSecurity',
    label: '安全',
    component: MdSecurity,
    category: IconCategory.Security,
    color: '#52c41a',
  },
  {
    name: 'FaShieldAlt',
    label: '防护',
    component: FaShieldAlt,
    category: IconCategory.Security,
    color: '#1890ff',
  },

  // 集成/连接类
  {
    name: 'BiNetworkChart',
    label: '网络',
    component: BiNetworkChart,
    category: IconCategory.Integration,
    color: '#1890ff',
  },
  {
    name: 'IoMdGitNetwork',
    label: '节点网络',
    component: IoMdGitNetwork,
    category: IconCategory.Integration,
    color: '#722ed1',
  },
  {
    name: 'BiTransfer',
    label: '数据传输',
    component: BiTransfer,
    category: IconCategory.Integration,
    color: '#13c2c2',
  },
  {
    name: 'AiOutlineCloudServer',
    label: '云服务',
    component: AiOutlineCloudServer,
    category: IconCategory.Integration,
    color: '#2f54eb',
  },

  // 媒体类
  {
    name: 'FaVideo',
    label: '视频',
    component: FaVideo,
    category: IconCategory.Media,
    color: '#eb2f96',
  },

  // 其他
  {
    name: 'MdSettings',
    label: '配置',
    component: MdSettings,
    category: IconCategory.Other,
    color: '#8c8c8c',
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
 * 根据图标名称获取默认颜色
 */
export const getIconColor = (iconName?: string): string | undefined => {
  if (!iconName) return undefined;
  const iconOption = ICON_OPTIONS.find((option) => option.name === iconName);
  return iconOption?.color;
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
