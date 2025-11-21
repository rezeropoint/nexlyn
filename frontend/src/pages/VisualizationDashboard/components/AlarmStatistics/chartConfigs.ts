import { getAlarmColor, getLineConfig as getThemeLineConfig, getPieConfig as getThemePieConfig } from "@/utils/antvTheme";
import type { GlobalToken } from "antd/es/theme/interface";
import type { AlarmStats } from "./types";

/**
 * 生成趋势图配置
 * @param trendData 趋势数据
 * @param token Ant Design token
 * @param isDark 是否为暗色主题
 */
export const getLineConfig = (
  trendData: Array<{ time: string; count: number }>,
  token: GlobalToken,
  isDark: boolean
) => {
  return getThemeLineConfig(trendData, "time", "count", {
    color: token.colorWarning,
    smooth: true,
    area: true,
    token,
    isDark,
  });
};

/**
 * 生成饼图配置
 * @param levelDistribution 告警级别分布数据
 * @param alarmStats 告警统计数据
 * @param token Ant Design token
 * @param isDark 是否为暗色主题
 */
export const getPieConfig = (
  levelDistribution: Array<{ level: string; count: number }>,
  alarmStats: AlarmStats,
  token: GlobalToken,
  isDark: boolean
) => {
  // 告警级别颜色映射
  const colorMap = {
    严重: getAlarmColor("严重"),
    重要: getAlarmColor("重要"),
    一般: getAlarmColor("一般"),
    轻微: getAlarmColor("轻微"),
  };

  return getThemePieConfig(levelDistribution, "count", "level", {
    innerRadius: 0.6,
    colorMap,
    token,
    isDark,
    statistic: {
      title: "总告警",
      content: String(alarmStats.total),
    },
  });
};
