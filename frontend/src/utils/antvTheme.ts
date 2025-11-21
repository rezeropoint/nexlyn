/**
 * AntV 图表主题配置工具
 *
 * 基于 AntV 官方色板设计，支持亮色/暗色主题自动适配
 *
 * 设计原则：
 * - 使用 AntV 官方分类色板保证视觉一致性
 * - 支持主题切换，自动适配文本和背景色
 * - 遵循前端编码规范，使用 theme token
 */

import type { GlobalToken } from 'antd/es/theme/interface';

/**
 * AntV 官方分类色板
 * 用于描述分类数据，色相分布均衡，相邻颜色明暗差异明显
 */
const CATEGORICAL_COLORS = [
  '#00CFE2', // 青色
  '#4E98FF', // 蓝色
  '#E49629', // 橙色
  '#CE6CFF', // 紫色
  '#8D7BFF', // 靛蓝
  '#53B81F', // 绿色
  '#CE9C29', // 金色
  '#EC5FB2', // 粉红
  '#0F8EB6', // 深青
  '#00A56E', // 翠绿
  '#A0AE02', // 黄绿
  '#B36AFD', // 亮紫
  '#237CBC', // 深蓝
  '#1BD468', // 亮绿
  '#CE8032', // 深橙
  '#FF7AFF', // 亮粉
  '#545FD3', // 深靛
  '#A3D50C', // 柠檬绿
  '#D1BE00', // 黄色
  '#FC8DD7', // 浅粉
];

/**
 * 告警级别色板（遵循通用告警色设计规范）
 */
const ALARM_COLORS = {
  严重: '#f5222d', // 红色 - 严重告警
  重要: '#fa8c16', // 橙色 - 重要告警
  一般: '#faad14', // 黄色 - 一般告警
  轻微: '#52c41a', // 绿色 - 轻微告警
};

/**
 * 事件状态色板
 */
const EVENT_STATUS_COLORS = {
  pending: '#faad14',    // 待处理 - 警告色
  processing: '#1890ff', // 处理中 - 信息色
  completed: '#52c41a',  // 已完成 - 成功色
  cancelled: '#d9d9d9',  // 已取消 - 灰色
};

/**
 * 获取图表通用配置
 * 自动适配亮色/暗色主题
 */
export const getChartTheme = (token: GlobalToken, isDark: boolean = false) => {
  return {
    // 颜色配置
    colors10: CATEGORICAL_COLORS.slice(0, 10),
    colors20: CATEGORICAL_COLORS,

    // 字体配置
    fontFamily: token.fontFamily,

    // 文本样式
    defaultColor: token.colorText,
    subColor: token.colorTextSecondary,

    // 背景和边框
    backgroundColor: 'transparent', // 使用透明背景，继承容器背景
    brandColor: token.colorPrimary,

    // 组件样式
    components: {
      axis: {
        common: {
          title: {
            style: {
              fill: token.colorText,
              fontSize: 12,
            },
          },
          label: {
            style: {
              fill: token.colorTextSecondary,
              fontSize: 11,
            },
          },
          line: {
            style: {
              stroke: token.colorBorder,
              lineWidth: 1,
            },
          },
          grid: {
            line: {
              style: {
                stroke: isDark ? token.colorBorderSecondary : token.colorFillQuaternary,
                lineDash: [4, 4],
              },
            },
          },
        },
      },
      legend: {
        common: {
          itemName: {
            style: {
              fill: token.colorText,
              fontSize: 12,
            },
          },
        },
      },
      tooltip: {
        domStyles: {
          'g2-tooltip': {
            backgroundColor: token.colorBgElevated,
            boxShadow: token.boxShadowSecondary,
            borderRadius: token.borderRadius,
          },
          'g2-tooltip-title': {
            color: token.colorText,
            fontSize: 12,
            fontWeight: 500,
          },
          'g2-tooltip-list-item': {
            color: token.colorTextSecondary,
            fontSize: 11,
          },
        },
      },
    },
  };
};

/**
 * 获取折线图通用配置
 */
export const getLineConfig = (
  data: Array<{ [key: string]: any }>,
  xField: string,
  yField: string,
  options?: {
    color?: string;
    smooth?: boolean;
    area?: boolean;
    token?: GlobalToken;
    isDark?: boolean;
  }
) => {
  const { color, smooth = true, area = false, token, isDark = false } = options || {};

  const baseConfig: any = {
    data,
    xField,
    yField,
    smooth,
    color: color || (token?.colorPrimary || '#1890ff'),
    point: {
      size: 4,
      shape: 'circle',
      style: {
        fill: color || (token?.colorPrimary || '#1890ff'),
        stroke: token?.colorBgContainer || '#fff',
        lineWidth: 2,
      },
    },
    line: {
      style: {
        lineWidth: 2,
      },
    },
  };

  if (area && token) {
    baseConfig.area = {
      style: {
        fill: `l(270) 0:${color || token.colorPrimary}00 1:${color || token.colorPrimary}20`,
      },
    };
  }

  if (token) {
    baseConfig.xAxis = {
      label: {
        style: {
          fill: token.colorTextSecondary,
          fontSize: 11,
        },
      },
      line: {
        style: {
          stroke: token.colorBorder,
        },
      },
    };

    baseConfig.yAxis = {
      label: {
        style: {
          fill: token.colorTextSecondary,
        },
      },
      grid: {
        line: {
          style: {
            stroke: isDark ? token.colorBorderSecondary : token.colorFillQuaternary,
            lineDash: [4, 4],
          },
        },
      },
    };
  }

  return baseConfig;
};

/**
 * 获取饼图通用配置
 */
export const getPieConfig = (
  data: Array<{ [key: string]: any }>,
  angleField: string,
  colorField: string,
  options?: {
    innerRadius?: number;
    colors?: string[];
    colorMap?: Record<string, string>;
    token?: GlobalToken;
    isDark?: boolean;
    statistic?: {
      title?: string | false;
      content?: string;
    };
  }
) => {
  const {
    innerRadius = 0.6,
    colors,
    colorMap,
    token,
    statistic,
  } = options || {};

  const config: any = {
    data,
    angleField,
    colorField,
    radius: 0.8,
    innerRadius,
    legend: {
      position: 'bottom' as const,
      itemName: {
        style: {
          fill: token?.colorText || '#000',
          fontSize: 11,
        },
      },
    },
    label: {
      text: colorField,
      position: 'spider',
      style: {
        fill: token?.colorText || '#000',
        fontSize: 11,
      },
    },
    interactions: [{ type: 'element-active' }],
  };

  // 颜色配置
  if (colorMap) {
    config.color = (datum: any) => colorMap[datum[colorField]] || CATEGORICAL_COLORS[0];
  } else if (colors) {
    config.color = colors;
  } else {
    config.color = CATEGORICAL_COLORS;
  }

  // 环形图中心统计文本
  if (statistic && token) {
    config.statistic = {
      title: statistic.title !== false
        ? {
            style: {
              fontSize: 12,
              color: token.colorTextSecondary,
            },
            content: statistic.title || '总计',
          }
        : false,
      content: statistic.content
        ? {
            style: {
              fontSize: 18,
              color: token.colorPrimary,
              fontWeight: 'bold',
            },
            content: statistic.content,
          }
        : undefined,
    };
  }

  return config;
};

/**
 * 获取柱状图通用配置
 */
export const getColumnConfig = (
  data: Array<{ [key: string]: any }>,
  xField: string,
  yField: string,
  options?: {
    color?: string | string[];
    colorField?: string;
    colorMap?: Record<string, string>;
    token?: GlobalToken;
    isDark?: boolean;
  }
) => {
  const { color, colorField, colorMap, token, isDark = false } = options || {};

  const config: any = {
    data,
    xField,
    yField,
    columnStyle: {
      radius: [4, 4, 0, 0],
    },
  };

  // 颜色配置
  if (colorField && colorMap) {
    config.colorField = colorField;
    config.color = (datum: any) => colorMap[datum[colorField]] || CATEGORICAL_COLORS[0];
  } else if (color) {
    config.color = color;
  } else if (token) {
    config.color = token.colorPrimary;
  }

  if (token) {
    config.xAxis = {
      label: {
        style: {
          fill: token.colorTextSecondary,
          fontSize: 11,
        },
      },
    };

    config.yAxis = {
      label: {
        style: {
          fill: token.colorTextSecondary,
        },
      },
      grid: {
        line: {
          style: {
            stroke: isDark ? token.colorBorderSecondary : token.colorFillQuaternary,
            lineDash: [4, 4],
          },
        },
      },
    };
  }

  return config;
};

/**
 * 获取告警级别颜色
 */
export const getAlarmColor = (level: string): string => {
  return ALARM_COLORS[level as keyof typeof ALARM_COLORS] || ALARM_COLORS.一般;
};

/**
 * 获取事件状态颜色
 */
export const getEventStatusColor = (status: string): string => {
  return EVENT_STATUS_COLORS[status as keyof typeof EVENT_STATUS_COLORS] || EVENT_STATUS_COLORS.pending;
};

/**
 * 事件统计专用颜色（从 AntV 分类色板中选取）
 */
export const EVENT_CHART_COLORS = {
  trend: CATEGORICAL_COLORS[0],      // 趋势图：青色 #00CFE2
  pending: CATEGORICAL_COLORS[2],    // 待处理：橙色 #E49629
  processing: CATEGORICAL_COLORS[1], // 处理中：蓝色 #4E98FF
  completed: CATEGORICAL_COLORS[5],  // 已完成：绿色 #53B81F
};

/**
 * 导出颜色常量供其他组件使用
 */
export { CATEGORICAL_COLORS, ALARM_COLORS, EVENT_STATUS_COLORS };
