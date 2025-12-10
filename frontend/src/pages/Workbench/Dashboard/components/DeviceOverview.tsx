import * as AntdIcons from '@ant-design/icons';
import {
  ApartmentOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  QuestionCircleOutlined,
  TagsOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Pie } from '@ant-design/plots';
import type { RadioChangeEvent } from 'antd';
import { Card, Empty, Progress, Radio, Space, Spin, theme } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useModel } from 'umi';
import {
  getUnifiedDeviceStatistics,
  type UnifiedDeviceStatisticsData,
} from '@/services/device';
import { CATEGORICAL_COLORS, getPieConfig } from '@/utils/antvTheme';
import { useApp } from '@/utils/appContext';
import { getDeviceCategoryInfo } from '@/utils/deviceCategory';
import styles from './DeviceOverview.less';

export interface DeviceOverviewProps {
  orgId?: string;
  loading?: boolean;
  refreshTrigger?: number;
}

type GroupByType = 'none' | 'org' | 'tag';

const DeviceOverview: React.FC<DeviceOverviewProps> = ({
  orgId,
  refreshTrigger,
}) => {
  const { token } = theme.useToken();
  const { message } = useApp();
  const { initialState } = useModel('@@initialState');

  const [loading, setLoading] = useState(false);
  const [statistics, setStatistics] =
    useState<UnifiedDeviceStatisticsData | null>(null);
  const [groupBy, setGroupBy] = useState<GroupByType>('none');

  // 判断是否为暗色主题
  const isDarkTheme = useMemo(
    () =>
      initialState?.themeMode === 'dark' ||
      initialState?.themeMode === 'dark-compact',
    [initialState?.themeMode],
  );

  const fetchStatistics = useCallback(async () => {
    if (!orgId) return;

    setLoading(true);
    try {
      const data = await getUnifiedDeviceStatistics({
        organizationId: orgId,
        deviceType: 'all',
        groupBy: groupBy === 'none' ? undefined : groupBy,
      });
      setStatistics(data);
    } catch (error) {
      console.error('Failed to fetch device statistics:', error);
      message.error('获取设备统计数据失败');
    } finally {
      setLoading(false);
    }
  }, [orgId, groupBy, message, refreshTrigger]);

  useEffect(() => {
    fetchStatistics();
  }, [fetchStatistics]);

  const handleGroupByChange = (e: RadioChangeEvent) => {
    setGroupBy(e.target.value);
    // 不清空数据，避免切换时内容区域高度突变
  };

  // 计算统一在线率（设备 + 监控点）
  const unifiedOnlineRate = useMemo(() => {
    if (!statistics) return 0;
    const totalCount = statistics.deviceTotal + (statistics.channelTotal || 0);
    const onlineCount =
      statistics.deviceOnline + (statistics.channelOnline || 0);
    return totalCount > 0 ? Math.round((onlineCount / totalCount) * 100) : 0;
  }, [statistics]);

  // 准备饼图数据
  const pieData = useMemo(() => {
    if (!statistics) return [];

    if (groupBy === 'none') {
      // 类别分布
      if (!statistics.byCategory) return [];
      return Object.entries(statistics.byCategory)
        .filter(([_, stats]) => stats.total > 0)
        .map(([category, stats]) => {
          const info = getDeviceCategoryInfo(category);
          return {
            type: info.name,
            value: stats.total,
            rawType: category, // 用于图标查找
          };
        })
        .sort((a, b) => b.value - a.value);
    } else {
      // 分组分布
      if (!statistics.groups) return [];
      return statistics.groups.map((group) => ({
        type: group.groupName,
        value: group.deviceTotal,
      }));
    }
  }, [statistics, groupBy]);

  // 饼图配置
  const pieConfig = useMemo(() => {
    const total = pieData.reduce((sum, item) => sum + item.value, 0);
    const config = getPieConfig(pieData, 'value', 'type', {
      innerRadius: 0.6,
      colors: CATEGORICAL_COLORS,
      token,
      isDark: isDarkTheme,
      statistic: {
        title: false,
        content: String(total),
      },
    });
    return {
      ...config,
      legend: false, // 自定义图例或不显示
    };
  }, [pieData, token, isDarkTheme]);

  const renderContent = () => {
    if (!statistics) return <Empty description="暂无数据" />;

    const hasChannelData = (statistics.channelTotal || 0) > 0;

    return (
      <>
        {/* 顶部核心指标 */}
        <div className={styles.overviewSection}>
          {/* IoT 设备统计 */}
          <div className={`${styles.metricItem} ${styles.total}`}>
            <div className={styles.metricValue}>{statistics.deviceTotal}</div>
            <div className={styles.metricLabel}>总设备</div>
          </div>
          <div className={`${styles.metricItem} ${styles.online}`}>
            <div className={styles.metricValue}>{statistics.deviceOnline}</div>
            <div className={styles.metricLabel}>在线</div>
          </div>
          <div className={`${styles.metricItem} ${styles.offline}`}>
            <div className={styles.metricValue}>{statistics.deviceOffline}</div>
            <div className={styles.metricLabel}>离线</div>
          </div>

          {/* 监控点统计（如有） */}
          {hasChannelData && (
            <>
              <div className={styles.metricDivider} />
              <div className={`${styles.metricItem} ${styles.channel}`}>
                <div className={styles.metricValue}>
                  {statistics.channelTotal}
                </div>
                <div className={styles.metricLabel}>总监控点</div>
              </div>
              <div className={`${styles.metricItem} ${styles.channelOnline}`}>
                <div className={styles.metricValue}>
                  {statistics.channelOnline}
                </div>
                <div className={styles.metricLabel}>监控点在线</div>
              </div>
              <div className={`${styles.metricItem} ${styles.channelOffline}`}>
                <div className={styles.metricValue}>
                  {statistics.channelOffline}
                </div>
                <div className={styles.metricLabel}>监控点离线</div>
              </div>
            </>
          )}

          {/* 在线率 */}
          <div className={styles.metricDivider} />
          <div className={`${styles.metricItem} ${styles.rate}`}>
            <div className={styles.metricValue}>{unifiedOnlineRate}%</div>
            <div className={styles.metricLabel}>在线率</div>
          </div>
        </div>

        <div className={styles.radioGroup}>
          <Radio.Group
            value={groupBy}
            onChange={handleGroupByChange}
            buttonStyle="solid"
          >
            <Radio.Button value="none">设备类别</Radio.Button>
            <Radio.Button value="org">所属组织</Radio.Button>
            <Radio.Button value="tag">设备标签</Radio.Button>
          </Radio.Group>
        </div>

        {/* 图表与列表区域 */}
        <div className={styles.contentSection}>
          <div className={styles.chartColumn}>
            {pieData.length > 0 ? (
              <>
                <div className={styles.chartWrapper}>
                  <Pie {...pieConfig} />
                </div>
                <div className={styles.chartLegend}>
                  <Space
                    wrap
                    size={[8, 8]}
                    style={{ justifyContent: 'center' }}
                  >
                    {pieData.slice(0, 5).map((item, index) => (
                      <Space
                        key={item.type}
                        size={4}
                        className={styles.legendItem}
                      >
                        <span
                          className={styles.legendDot}
                          style={{
                            background:
                              CATEGORICAL_COLORS[
                                index % CATEGORICAL_COLORS.length
                              ],
                          }}
                        />
                        <span className="text-secondary">{item.type}</span>
                        <span className={styles.legendValue}>{item.value}</span>
                      </Space>
                    ))}
                    {pieData.length > 5 && (
                      <span className="text-tertiary" style={{ fontSize: 12 }}>
                        ...
                      </span>
                    )}
                  </Space>
                </div>
              </>
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无分布数据"
              />
            )}
          </div>

          <div className={styles.listColumn}>
            {groupBy === 'none'
              ? // 类别列表
                pieData.map((item) => {
                  const rawType = (item as any).rawType;
                  const categoryStats = statistics.byCategory?.[rawType];
                  const categoryInfo = getDeviceCategoryInfo(rawType);
                  // 动态获取图标组件
                  const IconComponent =
                    (AntdIcons as any)[categoryInfo.icon] ||
                    QuestionCircleOutlined;

                  if (!categoryStats) return null;

                  const rate =
                    categoryStats.total > 0
                      ? Math.round(
                          (categoryStats.online / categoryStats.total) * 100,
                        )
                      : 0;

                  return (
                    <div key={item.type} className={styles.listItem}>
                      <div
                        className={styles.itemIcon}
                        style={{ color: categoryInfo.color }}
                      >
                        <IconComponent />
                      </div>
                      <div className={styles.itemContent}>
                        <div className={styles.itemHeader}>
                          <span className={styles.itemName}>{item.type}</span>
                          <span className={styles.itemTotal}>
                            {categoryStats.total}台
                          </span>
                        </div>
                        <div className={styles.itemStats}>
                          <span className="online">
                            <CheckCircleOutlined /> {categoryStats.online}
                          </span>
                          <span className="offline">
                            <CloseCircleOutlined /> {categoryStats.offline}
                          </span>
                          <span>
                            <ThunderboltOutlined /> {rate}%
                          </span>
                        </div>
                        <Progress
                          percent={rate}
                          size="small"
                          showInfo={false}
                          strokeColor={categoryInfo.color}
                          trailColor={token.colorFillQuaternary}
                        />
                      </div>
                    </div>
                  );
                })
              : // 分组列表
                statistics.groups?.map((group) => {
                  const total = group.deviceTotal + (group.channelTotal || 0);
                  const online =
                    group.deviceOnline + (group.channelOnline || 0);
                  const rate =
                    total > 0 ? Math.round((online / total) * 100) : 0;
                  const hasChannel = (group.channelTotal || 0) > 0;

                  return (
                    <div key={group.groupKey} className={styles.listItem}>
                      <div className={`icon-primary ${styles.itemIcon}`}>
                        {groupBy === 'org' ? (
                          <ApartmentOutlined />
                        ) : (
                          <TagsOutlined />
                        )}
                      </div>
                      <div className={styles.itemContent}>
                        <div className={styles.itemHeader}>
                          <span
                            className={styles.itemName}
                            title={group.groupName}
                          >
                            {group.groupName}
                          </span>
                          <span className={styles.itemTotal}>
                            {group.deviceTotal}台
                            {hasChannel && ` + ${group.channelTotal}点`}
                          </span>
                        </div>
                        <div className={styles.itemStats}>
                          <span className="online">
                            <CheckCircleOutlined /> {online}
                          </span>
                          <span className="offline">
                            <CloseCircleOutlined /> {total - online}
                          </span>
                          <span>
                            <ThunderboltOutlined /> {rate}%
                          </span>
                        </div>
                        <Progress
                          percent={rate}
                          size="small"
                          showInfo={false}
                          strokeColor={token.colorPrimary}
                          trailColor={token.colorFillQuaternary}
                        />
                      </div>
                    </div>
                  );
                })}
          </div>
        </div>
      </>
    );
  };

  return (
    <Card
      className={styles.deviceOverviewCard}
      title={
        <Space>
          <AppstoreOutlined className="icon-primary" />
          <span>设备运行概览</span>
        </Space>
      }
      extra={
        <span className="text-tertiary" style={{ fontSize: 12 }}>
          实时监控
        </span>
      }
      loading={loading && !statistics} // Initial load
    >
      <Spin spinning={loading && !!statistics}>{renderContent()}</Spin>
    </Card>
  );
};

export default DeviceOverview;
