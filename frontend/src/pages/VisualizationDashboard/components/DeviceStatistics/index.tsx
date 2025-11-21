import {
  type DeviceStatisticsGroup,
  getUnifiedDeviceStatistics,
  type UnifiedDeviceStatisticsData,
} from "@/services/device";
import { useApp } from "@/utils/appContext";
import { CATEGORICAL_COLORS, getPieConfig } from "@/utils/antvTheme";
import { getDeviceCategoryInfo } from "@/utils/deviceCategory";
import * as AntdIcons from "@ant-design/icons";
import {
  ApartmentOutlined,
  ApiOutlined,
  AppstoreOutlined,
  QuestionCircleOutlined,
  TagsOutlined,
} from "@ant-design/icons";
import { Pie } from "@ant-design/plots";
import { useModel } from "@umijs/max";
import type { RadioChangeEvent } from "antd";
import { Card, Empty, Progress, Radio, Space, Spin, theme } from "antd";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import "./index.less";

export interface DeviceStatisticsProps {
  className?: string;
  orgId?: string; // 组织ID（从父组件传入）
}

type GroupByType = "none" | "org" | "tag";

const DeviceStatistics: React.FC<DeviceStatisticsProps> = ({
  className,
  orgId,
}) => {
  const { message } = useApp();
  const { initialState } = useModel("@@initialState");
  const { token } = theme.useToken();
  const [loading, setLoading] = useState(false);
  const [statistics, setStatistics] =
    useState<UnifiedDeviceStatisticsData | null>(null);
  const [hasError, setHasError] = useState(false);
  const [groupBy, setGroupBy] = useState<GroupByType>("none");

  // 判断是否为暗色主题（基于themeMode）
  const isDarkTheme = useMemo(
    () =>
      initialState?.themeMode === "dark" ||
      initialState?.themeMode === "dark-compact",
    [initialState?.themeMode]
  );

  // 加载设备统计数据
  const fetchStatistics = useCallback(async () => {
    if (!orgId) {
      return;
    }

    setLoading(true);
    setHasError(false);
    try {
      const data = await getUnifiedDeviceStatistics({
        organizationId: orgId,
        deviceType: "all", // 始终显示全部设备
        groupBy: groupBy === "none" ? undefined : groupBy,
      });

      setStatistics(data);
    } catch (_error) {
      setHasError(true);
      message.error("获取设备统计失败");
    } finally {
      setLoading(false);
    }
  }, [orgId, groupBy]);

  useEffect(() => {
    fetchStatistics();
  }, [fetchStatistics]);

  // 处理分组切换
  const handleGroupByChange = useCallback((e: RadioChangeEvent) => {
    const newGroupBy = e.target.value as GroupByType;
    setGroupBy(newGroupBy);
    // 立即清空数据和设置loading状态，避免显示旧数据
    setStatistics(null);
    setLoading(true);
  }, []);

  // 统一计算设备和通道的在线率
  const unifiedOnlineRate = useMemo(() => {
    if (!statistics) return "0.0";

    const totalCount = statistics.deviceTotal + (statistics.channelTotal || 0);
    const onlineCount =
      statistics.deviceOnline + (statistics.channelOnline || 0);

    return totalCount > 0
      ? ((onlineCount / totalCount) * 100).toFixed(1)
      : "0.0";
  }, [statistics]);

  // 判断是否有通道数据
  const hasChannelData = statistics && (statistics.channelTotal || 0) > 0;

  // 准备饼图数据（用于分组统计）
  const pieData = useMemo(() => {
    if (!statistics?.groups || statistics.groups.length === 0) {
      return [];
    }
    return statistics.groups.map((group: DeviceStatisticsGroup) => ({
      type: group.groupName,
      value: group.deviceTotal,
    }));
  }, [statistics?.groups]);

  // 准备设备类别饼图数据（用于总览模式）
  const categoryPieData = useMemo(() => {
    if (!statistics?.byCategory) {
      return [];
    }
    return Object.entries(statistics.byCategory)
      .filter(([_, stats]) => stats.total > 0)
      .map(([category, stats]) => {
        const categoryInfo = getDeviceCategoryInfo(category);
        return {
          type: categoryInfo.name,
          value: stats.total,
        };
      });
  }, [statistics?.byCategory]);

  // 分组饼图配置
  const groupPieConfig = useMemo(() => {
    const totalCount = pieData.reduce((sum, item) => sum + item.value, 0);
    return getPieConfig(pieData, "value", "type", {
      innerRadius: 0.6,
      colors: CATEGORICAL_COLORS,
      token,
      isDark: isDarkTheme,
      statistic: {
        title: false,
        content: String(totalCount),
      },
    });
  }, [pieData, token, isDarkTheme]);

  // 类别饼图配置
  const categoryPieConfig = useMemo(() => {
    const totalCount = categoryPieData.reduce(
      (sum, item) => sum + item.value,
      0
    );
    return getPieConfig(categoryPieData, "value", "type", {
      innerRadius: 0.6,
      colors: CATEGORICAL_COLORS,
      token,
      isDark: isDarkTheme,
      statistic: {
        title: false,
        content: String(totalCount),
      },
    });
  }, [categoryPieData, token, isDarkTheme]);

  // 显示空数据状态
  if (!loading && (hasError || !statistics)) {
    return (
      <div className={`device-statistics ${className || ""}`}>
        <Card className="empty-card" size="small">
          <Empty
            description={hasError ? "未能获取到数据" : "暂无设备数据"}
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      </div>
    );
  }

  return (
    <Spin spinning={loading}>
      <div className={`device-statistics ${className || ""}`}>
        {/* 总体统计卡片 */}
        <Card
          className="overview-card"
          title={
            <Space>
              <AppstoreOutlined />
              <span>设备统计</span>
            </Space>
          }
          extra={
            <Radio.Group
              value={groupBy}
              onChange={handleGroupByChange}
              buttonStyle="solid"
              size="small"
            >
              <Radio.Button value="none">总览</Radio.Button>
              <Radio.Button value="org">
                <ApartmentOutlined />
              </Radio.Button>
              <Radio.Button value="tag">
                <TagsOutlined />
              </Radio.Button>
            </Radio.Group>
          }
          size="small"
        >
          {/* 统一的设备和通道统计 */}
          <div className="stats-row-unified">
            <div className="stat-item total">
              <div className="stat-value">{statistics?.deviceTotal || 0}</div>
              <div className="stat-label">总设备</div>
            </div>
            <div className="stat-item online">
              <div className="stat-value">{statistics?.deviceOnline || 0}</div>
              <div className="stat-label">在线</div>
            </div>
            <div className="stat-item offline">
              <div className="stat-value">{statistics?.deviceOffline || 0}</div>
              <div className="stat-label">离线</div>
            </div>

            {hasChannelData && (
              <>
                <div className="vertical-divider"></div>
                <div className="stat-item total channel">
                  <div className="stat-value">
                    {statistics?.channelTotal || 0}
                  </div>
                  <div className="stat-label">总监控点</div>
                </div>
                <div className="stat-item online channel">
                  <div className="stat-value">
                    {statistics?.channelOnline || 0}
                  </div>
                  <div className="stat-label">监控点在线</div>
                </div>
                <div className="stat-item offline channel">
                  <div className="stat-value">
                    {statistics?.channelOffline || 0}
                  </div>
                  <div className="stat-label">监控点离线</div>
                </div>
              </>
            )}
          </div>

          {/* 分割线 */}
          <div className="horizontal-divider"></div>

          {/* 统一在线率进度条 */}
          <div className="progress-container">
            <div className="progress-item">
              <div className="progress-header">
                <span className="progress-label">
                  在线率
                  {hasChannelData && (
                    <span className="progress-detail">
                      （{statistics?.deviceOnline}台设备 +{" "}
                      {statistics?.channelOnline}个监控点）
                    </span>
                  )}
                </span>
                <span className="progress-value">{unifiedOnlineRate}%</span>
              </div>
              <Progress
                percent={parseFloat(unifiedOnlineRate)}
                strokeColor={{
                  "0%": token.colorSuccess,
                  "100%": token.colorPrimary,
                }}
                trailColor={token.colorFillQuaternary}
                showInfo={false}
                strokeLinecap="round"
                size={["100%", 8]}
              />
            </div>
          </div>
        </Card>

        {/* 分组统计列表 */}
        {statistics?.groups && statistics.groups.length > 0 && (
          <Card
            key={`groups-${groupBy}`}
            className="groups-card"
            title={
              <Space>
                {groupBy === "org" ? <ApartmentOutlined /> : <TagsOutlined />}
                <span>分组统计</span>
              </Space>
            }
            size="small"
          >
            {/* 饼图展示比例 */}
            {pieData.length > 0 && (
              <div
                key={`pie-${groupBy}-${isDarkTheme ? "dark" : "light"}`}
                className="groups-chart"
              >
                <Pie
                  key={`pie-chart-${groupBy}-${isDarkTheme ? "dark" : "light"}`}
                  {...groupPieConfig}
                  height={200}
                />
              </div>
            )}

            <div className="groups-list">
              {statistics.groups.map((group: DeviceStatisticsGroup) => {
                // 统一计算设备和通道的在线率
                const totalCount =
                  group.deviceTotal + (group.channelTotal || 0);
                const onlineCount =
                  group.deviceOnline + (group.channelOnline || 0);
                const groupUnifiedOnlineRate =
                  totalCount > 0
                    ? ((onlineCount / totalCount) * 100).toFixed(0)
                    : "0";

                const hasGroupChannelData = (group.channelTotal || 0) > 0;

                return (
                  <div key={group.groupKey} className="group-item">
                    <div className="group-header">
                      <span className="group-name" title={group.groupName}>
                        {group.groupName}
                      </span>
                      <span className="group-summary">
                        {group.deviceTotal}台设备
                        {hasGroupChannelData &&
                          ` | ${group.channelTotal}个监控点`}
                      </span>
                    </div>

                    {/* 统一的设备和通道统计 */}
                    <div className="group-stats-unified">
                      <div className="group-stat online">
                        <span className="label">设备在线</span>
                        <span className="value">{group.deviceOnline}</span>
                      </div>
                      <div className="group-stat offline">
                        <span className="label">设备离线</span>
                        <span className="value">{group.deviceOffline}</span>
                      </div>

                      {hasGroupChannelData && (
                        <>
                          <div className="stat-divider"></div>
                          <div className="group-stat online channel">
                            <span className="label">监控点在线</span>
                            <span className="value">{group.channelOnline}</span>
                          </div>
                          <div className="group-stat offline channel">
                            <span className="label">监控点离线</span>
                            <span className="value">
                              {group.channelOffline}
                            </span>
                          </div>
                        </>
                      )}
                    </div>

                    {/* 统一在线率进度条 */}
                    <div className="group-progress-container">
                      <div className="progress-row">
                        <span className="progress-label">在线率</span>
                        <Progress
                          percent={parseFloat(groupUnifiedOnlineRate)}
                          strokeColor={{
                            "0%": token.colorSuccess,
                            "100%": token.colorPrimary,
                          }}
                          trailColor={token.colorFillQuaternary}
                          showInfo={false}
                          strokeLinecap="round"
                          size={["100%", 4]}
                        />
                        <span className="progress-percent">
                          {groupUnifiedOnlineRate}%
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </Card>
        )}

        {/* 无分组数据时的提示 */}
        {groupBy !== "none" &&
          (!statistics?.groups || statistics.groups.length === 0) && (
            <Card
              key={`empty-${groupBy}`}
              className="empty-groups-card"
              size="small"
            >
              <Empty
                description={
                  groupBy === "org" ? "暂无组织分组数据" : "暂无标签分组数据"
                }
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            </Card>
          )}

        {/* 按设备类别统计（仅在无其他分组时显示） */}
        {groupBy === "none" &&
          statistics?.byCategory &&
          (() => {
            const categoryList = Object.entries(statistics.byCategory)
              .filter(([_, stats]) => stats.total > 0) // 只显示有设备的类别
              .sort((a, b) => b[1].total - a[1].total); // 按设备数量降序排序

            return (
              <Card
                key="type-stats"
                className="type-stats-card"
                title={
                  <Space>
                    <ApiOutlined />
                    <span>设备类别分布</span>
                  </Space>
                }
                size="small"
              >
                {categoryList.length > 0 ? (
                  <>
                    {/* 类别分布饼图 */}
                    {categoryPieData.length > 0 && (
                      <div
                        key={`category-pie-${isDarkTheme ? "dark" : "light"}`}
                        className="category-chart"
                      >
                        <Pie
                          key={`category-chart-${
                            isDarkTheme ? "dark" : "light"
                          }`}
                          {...categoryPieConfig}
                          height={200}
                        />
                      </div>
                    )}

                    <div className="type-stats-list">
                      {categoryList.map(([category, stats]) => {
                        const categoryInfo = getDeviceCategoryInfo(category);
                        const IconComponent =
                          (
                            AntdIcons as unknown as Record<
                              string,
                              React.ComponentType<{
                                style?: React.CSSProperties;
                              }>
                            >
                          )[categoryInfo.icon] || QuestionCircleOutlined;
                        const onlineRate =
                          stats.total > 0
                            ? parseFloat(
                                ((stats.online / stats.total) * 100).toFixed(0)
                              )
                            : 0;

                        return (
                          <div key={category} className="type-stat-item">
                            <div className="type-header">
                              <IconComponent
                                style={{
                                  marginRight: 8,
                                  color: categoryInfo.color,
                                }}
                              />
                              <span className="type-name">
                                {categoryInfo.name}
                              </span>
                              <span className="type-total">
                                {stats.total}台
                              </span>
                            </div>
                            <div className="type-detail">
                              <span className="online-count">
                                在线: {stats.online}
                              </span>
                              <span className="offline-count">
                                离线: {stats.offline}
                              </span>
                            </div>
                            <Progress
                              percent={onlineRate}
                              strokeColor={categoryInfo.color}
                              trailColor={token.colorFillQuaternary}
                              showInfo={false}
                              strokeLinecap="round"
                              size={["100%", 4]}
                            />
                          </div>
                        );
                      })}
                    </div>
                  </>
                ) : (
                  <Empty
                    description="暂无设备类别数据"
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                  />
                )}
              </Card>
            );
          })()}
      </div>
    </Spin>
  );
};

export default DeviceStatistics;
