import type { ColumnInfo } from "@/pages/EventManagement/types";
import {
  getDurationStats,
  getEventDetail,
  getPendingStats,
  getTrendStats,
  queryEventData,
} from "@/services/eventhandler";
import type {
  DurationStatsData,
  PendingStatsData,
  TrendStatsData,
} from "@/services/eventhandler/types";
import {
  EVENT_CHART_COLORS,
  getLineConfig,
  getPieConfig,
} from "@/utils/antvTheme";
import { Line, Pie } from "@ant-design/plots";
import {
  BarChartOutlined,
  CalendarOutlined,
  CheckOutlined,
  ExclamationCircleOutlined,
  FieldTimeOutlined,
  LeftOutlined,
  LineChartOutlined,
  LoadingOutlined,
  PieChartOutlined,
  UnorderedListOutlined,
} from "@ant-design/icons";
import { useModel } from "@umijs/max";
import {
  Button,
  Card,
  DatePicker,
  Empty,
  Flex,
  message,
  Pagination,
  Space,
  Spin,
  Statistic,
  theme,
} from "antd";
import dayjs, { type Dayjs } from "dayjs";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import EventDetailContent, { type EventDetailData } from "./EventDetailContent";
import "./index.less";

const { RangePicker } = DatePicker;

// 定义事件记录接口
interface EventRecord {
  slp_journey_id: number | string;
  _rowKey: string;
  [key: string]: unknown;
}

// 使用 EventManagement 的 ColumnInfo 类型
type ColumnConfig = ColumnInfo;

export interface EventStatisticsProps {
  className?: string;
  orgId?: string; // 组织ID（从父组件传入）
}

const EventStatistics: React.FC<EventStatisticsProps> = ({
  className,
  orgId,
}) => {
  const { token } = theme.useToken();
  const { initialState } = useModel("@@initialState");
  const [loading, setLoading] = useState(false);
  const [hasError, setHasError] = useState(false);

  // 判断是否为暗色主题
  const isDark = useMemo(
    () =>
      initialState?.themeMode === "dark" ||
      initialState?.themeMode === "dark-compact",
    [initialState?.themeMode]
  );

  // 时间范围状态（默认最近30天）
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([
    dayjs().subtract(30, "day"),
    dayjs(),
  ]);

  // 统计数据
  const [pendingStats, setPendingStats] = useState<PendingStatsData | null>(
    null
  );
  const [durationStats, setDurationStats] = useState<DurationStatsData | null>(
    null
  );
  const [trendStats, setTrendStats] = useState<TrendStatsData | null>(null);

  // 列表视图状态
  const [showList, setShowList] = useState(false);
  const [selectedEventId, setSelectedEventId] = useState<string>("");
  const [selectedEventName, setSelectedEventName] = useState<string>("");
  const [selectedStatus, setSelectedStatus] = useState<string[]>([]);
  const [listLoading, setListLoading] = useState(false);
  const [listData, setListData] = useState<EventRecord[]>([]);
  const [listTotal, setListTotal] = useState(0);
  const [listPage, setListPage] = useState(1);
  const [listPageSize] = useState(3);
  const [backendColumns, setBackendColumns] = useState<ColumnConfig[]>([]);

  // 详情视图状态
  const [showDetail, setShowDetail] = useState(false);
  const [currentJourneyId, setCurrentJourneyId] = useState<number | null>(null);
  const [detailData, setDetailData] = useState<EventDetailData | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);

  // 加载统计数据
  const loadStats = useCallback(async () => {
    if (!orgId) {
      return;
    }

    setLoading(true);
    setHasError(false);

    try {
      // 使用当前选择的时间范围
      const dateFrom = dateRange[0].toISOString();
      const dateTo = dateRange[1].toISOString();

      // 查询待处理事件
      const pendingParams = {
        eventIds: [], // 空数组表示查询所有事件
        orgId,
        dateFrom,
        dateTo,
      };

      // 查询时长统计
      const durationParams = {
        eventIds: [],
        orgId,
        dateFrom,
        dateTo,
      };

      // 查询趋势统计
      const trendParams = {
        eventIds: [],
        orgId,
        dateFrom,
        dateTo,
        granularity: "day" as const, // 按天统计
      };

      const [pendingRes, durationRes, trendRes] = await Promise.all([
        getPendingStats(pendingParams),
        getDurationStats(durationParams),
        getTrendStats(trendParams),
      ]);

      if (pendingRes.code === 0) {
        setPendingStats(pendingRes.data || null);
      }

      if (durationRes.code === 0) {
        setDurationStats(durationRes.data || null);
      }

      if (trendRes.code === 0) {
        setTrendStats(trendRes.data || null);
      }

      if (
        pendingRes.code !== 0 &&
        durationRes.code !== 0 &&
        trendRes.code !== 0
      ) {
        setHasError(true);
        message.error("未能获取到统计数据");
      }
    } catch (_error) {
      setHasError(true);
      message.error("加载统计数据失败");
    } finally {
      setLoading(false);
    }
  }, [orgId, dateRange]);

  // 当组织或时间范围变化时重新加载数据
  useEffect(() => {
    loadStats();
  }, [loadStats]);

  // 格式化时长（秒转为可读格式）
  const formatDuration = (seconds: number): string => {
    if (seconds < 60) return `${seconds.toFixed(0)}秒`;
    if (seconds < 3600) return `${(seconds / 60).toFixed(1)}分钟`;
    if (seconds < 86400) return `${(seconds / 3600).toFixed(1)}小时`;
    return `${(seconds / 86400).toFixed(1)}天`;
  };

  // 趋势图数据
  const trendChartData = useMemo(() => {
    if (!trendStats?.timePoints || trendStats.timePoints.length === 0) {
      return [];
    }
    return trendStats.timePoints.map((item) => ({
      date: dayjs(item.date).format("MM-DD"),
      count: item.totalCount,
    }));
  }, [trendStats]);

  // 状态分布饼图数据
  const statusPieData = useMemo(() => {
    if (!pendingStats) return [];

    const data = [];
    if (pendingStats.totalPending > 0) {
      data.push({ status: "待处理", count: pendingStats.totalPending });
    }
    if (pendingStats.totalProcessing > 0) {
      data.push({ status: "处理中", count: pendingStats.totalProcessing });
    }
    if (durationStats && durationStats.completedCount > 0) {
      data.push({ status: "已完成", count: durationStats.completedCount });
    }

    return data;
  }, [pendingStats, durationStats]);

  // 趋势图配置
  const trendLineConfig = useMemo(
    () =>
      getLineConfig(trendChartData, "date", "count", {
        color: EVENT_CHART_COLORS.trend,
        smooth: true,
        area: true,
        token,
        isDark,
      }),
    [trendChartData, token, isDark]
  );

  // 状态分布饼图配置
  const statusPieConfig = useMemo(() => {
    const colorMap = {
      待处理: EVENT_CHART_COLORS.pending,
      处理中: EVENT_CHART_COLORS.processing,
      已完成: EVENT_CHART_COLORS.completed,
    };

    const totalCount = statusPieData.reduce(
      (sum, item) => sum + item.count,
      0
    );

    return getPieConfig(statusPieData, "count", "status", {
      innerRadius: 0.6,
      colorMap,
      token,
      isDark,
      statistic: {
        title: "总计",
        content: String(totalCount),
      },
    });
  }, [statusPieData, token, isDark]);

  // 加载事件列表数据
  const loadEventList = useCallback(async () => {
    if (!selectedEventId || !orgId) {
      return;
    }

    setListLoading(true);
    try {
      const response = await queryEventData(selectedEventId, {
        eventId: selectedEventId,
        orgId,
        page: listPage,
        pageSize: listPageSize,
        status: selectedStatus,
      });

      if (response.code === 0 && response.data) {
        const records = (response.data.records || []).map(
          (record: Record<string, unknown>, index: number): EventRecord => ({
            ...record,
            slp_journey_id: record.slp_journey_id as number | string,
            _rowKey: `${listPage}_${index}`,
          })
        );
        setListData(records);
        setListTotal(response.data.total || 0);
        if (response.data.columns) {
          setBackendColumns(response.data.columns);
        }
      } else {
        message.error(response.msg || "查询失败");
      }
    } catch (_error) {
      message.error("查询失败");
    } finally {
      setListLoading(false);
    }
  }, [selectedEventId, orgId, listPage, listPageSize, selectedStatus]);

  // 查看详情（切换到详情视图）
  const handleViewDetail = useCallback(
    async (journeyId: number | string) => {
      if (!selectedEventId) {
        message.warning("缺少事件配置ID");
        return;
      }

      const numericJourneyId =
        typeof journeyId === "string" ? parseInt(journeyId, 10) : journeyId;
      setCurrentJourneyId(numericJourneyId);
      setShowDetail(true);
      setDetailLoading(true);
      setDetailData(null);

      try {
        const response = await getEventDetail(
          selectedEventId,
          numericJourneyId
        );

        if (response.code === 0 && response.data) {
          setDetailData(response.data);
        } else {
          message.error(response.msg || response.message || "获取详情失败");
        }
      } catch (_error) {
        message.error("获取详情失败");
      } finally {
        setDetailLoading(false);
      }
    },
    [selectedEventId]
  );

  // 返回列表视图
  const handleBackToList = () => {
    setShowDetail(false);
    setCurrentJourneyId(null);
    setDetailData(null);
  };

  // 处理点击事件统计项
  const handleClickEventStatus = (
    eventConfigID: string,
    eventName: string,
    status: string[]
  ) => {
    if (!orgId) {
      message.warning("请先选择组织");
      return;
    }
    setSelectedEventId(eventConfigID);
    setSelectedEventName(eventName);
    setSelectedStatus(status);
    setListPage(1); // 重置页码
    setShowList(true);
  };

  // 返回统计视图
  const handleBackToStats = () => {
    setShowList(false);
    setSelectedEventId("");
    setSelectedEventName("");
    setSelectedStatus([]);
    setListData([]);
    // 清除详情视图状态
    setShowDetail(false);
    setCurrentJourneyId(null);
    setDetailData(null);
  };

  // 当列表参数变化时加载数据
  useEffect(() => {
    if (showList && selectedEventId && orgId) {
      loadEventList();
    }
  }, [showList, selectedEventId, orgId, loadEventList]);

  // 显示空数据状态
  if (!loading && !orgId) {
    return (
      <div className={`event-statistics ${className || ""}`}>
        <Card className="empty-card">
          <Empty
            description="请选择组织"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      </div>
    );
  }

  if (!loading && hasError) {
    return (
      <div className={`event-statistics ${className || ""}`}>
        <Card className="empty-card">
          <Empty
            description="未能获取到数据"
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      </div>
    );
  }

  return (
    <Spin spinning={loading}>
      <div className={`event-statistics ${className || ""}`}>
        {/* 时间范围选择器 */}
        <Card variant="borderless" size="small" className="time-range-card">
          <Space
            direction="vertical"
            className="time-range-content"
            size="small"
          >
            <div className="time-range-header">
              <CalendarOutlined className="calendar-icon" />
              <span className="header-text">事件时间范围</span>
            </div>
            <RangePicker
              value={dateRange}
              onChange={(dates) => {
                if (dates?.[0] && dates[1]) {
                  setDateRange([dates[0], dates[1]]);
                }
              }}
              size="small"
              className="range-picker"
              format="YYYY-MM-DD"
              allowClear={false}
              presets={[
                {
                  label: "最近24小时",
                  value: [dayjs().subtract(1, "day"), dayjs()],
                },
                {
                  label: "最近3天",
                  value: [dayjs().subtract(3, "day"), dayjs()],
                },
                {
                  label: "最近7天",
                  value: [dayjs().subtract(7, "day"), dayjs()],
                },
                {
                  label: "最近30天",
                  value: [dayjs().subtract(30, "day"), dayjs()],
                },
              ]}
            />
          </Space>
        </Card>

        {/* 总体统计 */}
        <Card
          variant="borderless"
          className="total-stats-card"
          size="small"
          title={
            <Space>
              <BarChartOutlined />
              <span>事件统计</span>
            </Space>
          }
        >
          <Flex gap={8} wrap="nowrap" justify="space-between">
            <Statistic
              title="平均时长"
              value={durationStats?.avgDuration || 0}
              formatter={(value) => formatDuration(Number(value))}
              prefix={<FieldTimeOutlined className="stat-icon" />}
              valueStyle={{ color: token.colorPrimary, fontSize: "14px" }}
            />
            <Statistic
              title="已完成"
              value={durationStats?.completedCount || 0}
              prefix={<CheckOutlined className="stat-icon" />}
              suffix={`/${durationStats?.totalCount || 0}`}
              valueStyle={{ color: token.colorSuccess, fontSize: "14px" }}
            />
            <Statistic
              title="待处理"
              value={pendingStats?.totalPending || 0}
              prefix={<ExclamationCircleOutlined className="stat-icon" />}
              valueStyle={{ color: token.colorWarning, fontSize: "14px" }}
            />
            <Statistic
              title="处理中"
              value={pendingStats?.totalProcessing || 0}
              prefix={<LoadingOutlined className="stat-icon" />}
              valueStyle={{ color: token.colorLink, fontSize: "14px" }}
            />
          </Flex>
        </Card>

        {/* 事件趋势图 */}
        <Card
          variant="borderless"
          className="trend-chart-card"
          size="small"
          title={
            <Space>
              <LineChartOutlined />
              <span>事件趋势</span>
            </Space>
          }
        >
          {trendChartData.length > 0 ? (
            <Line {...trendLineConfig} height={120} />
          ) : (
            <Empty
              description="暂无趋势数据"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          )}
        </Card>

        {/* 状态分布图 */}
        <Card
          variant="borderless"
          className="status-chart-card"
          size="small"
          title={
            <Space>
              <PieChartOutlined />
              <span>状态分布</span>
            </Space>
          }
        >
          {statusPieData.length > 0 ? (
            <Pie {...statusPieConfig} height={140} />
          ) : (
            <Empty
              description="暂无状态数据"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          )}
        </Card>

        {/* 事件分类统计 */}
        <Card
          variant="borderless"
          className="event-list-card"
          size="small"
          title={
            <Space>
              <UnorderedListOutlined />
              <span>
                {!showList
                  ? "事件分类"
                  : !showDetail
                  ? `${selectedEventName} ${
                      selectedStatus.includes("pending") &&
                      selectedStatus.includes("processing")
                        ? "(全部)"
                        : selectedStatus.includes("pending")
                        ? "(待处理)"
                        : "(处理中)"
                    }`
                  : `事件详情 #${currentJourneyId}`}
              </span>
            </Space>
          }
          extra={
            showList && !showDetail ? (
              <Button
                type="text"
                size="small"
                icon={<LeftOutlined />}
                onClick={handleBackToStats}
                className="back-button"
              >
                返回
              </Button>
            ) : showDetail ? (
              <Button
                type="text"
                size="small"
                icon={<LeftOutlined />}
                onClick={handleBackToList}
                className="back-button"
              >
                返回列表
              </Button>
            ) : null
          }
        >
          {!showList ? (
            // ========== 统计视图 ==========
            pendingStats?.eventStats && pendingStats.eventStats.length > 0 ? (
              <div className="event-items">
                {pendingStats.eventStats.map((event) => (
                  <div key={event.eventConfigId} className="event-item">
                    <div className="event-name">{event.eventName}</div>
                    <div className="event-stats">
                      <div
                        className="stat-item pending"
                        onClick={() =>
                          handleClickEventStatus(
                            event.eventConfigId,
                            event.eventName,
                            ["pending"]
                          )
                        }
                      >
                        <div className="stat-label">待处理</div>
                        <div className="stat-value">{event.pendingCount}</div>
                      </div>
                      <div className="stat-divider" />
                      <div
                        className="stat-item processing"
                        onClick={() =>
                          handleClickEventStatus(
                            event.eventConfigId,
                            event.eventName,
                            ["processing"]
                          )
                        }
                      >
                        <div className="stat-label">处理中</div>
                        <div className="stat-value">
                          {event.processingCount}
                        </div>
                      </div>
                      <div className="stat-divider" />
                      <div
                        className="stat-item total"
                        onClick={() =>
                          handleClickEventStatus(
                            event.eventConfigId,
                            event.eventName,
                            ["pending", "processing"]
                          )
                        }
                      >
                        <div className="stat-label">小计</div>
                        <div className="stat-value">{event.total}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <Empty
                description="暂无待处理事件"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )
          ) : !showDetail ? (
            // ========== 列表视图 ==========
            <div className="event-list-view">
              {listLoading ? (
                <div className="list-loading">
                  <Spin size="large" />
                </div>
              ) : listData.length > 0 ? (
                <>
                  <div className="list-items">
                    {listData.map((record) => {
                      const journeyId = record.slp_journey_id;
                      const displayFields = backendColumns.filter(
                        (col) => col.field !== "slp_journey_id"
                      );

                      return (
                        <div key={record._rowKey} className="list-item">
                          {/* Journey ID + 详情按钮 */}
                          <div className="item-header">
                            <div className="journey-info">
                              <span className="journey-label">事件ID:</span>
                              <span className="journey-id">{journeyId}</span>
                            </div>
                            <Button
                              type="link"
                              size="small"
                              onClick={() => handleViewDetail(journeyId)}
                              className="detail-button"
                            >
                              详情
                            </Button>
                          </div>

                          {/* 其他关键字段 */}
                          <div className="item-fields">
                            {displayFields.map((col) => (
                              <div key={col.field} className="field-row">
                                <span className="field-label">
                                  {col.displayName || col.field}:
                                </span>
                                <span
                                  className="field-value"
                                  title={String(record[col.field] || "-")}
                                >
                                  {String(record[col.field] || "-")}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  {/* 分页器 */}
                  {listTotal > listPageSize && (
                    <div className="list-pagination">
                      <Pagination
                        simple
                        current={listPage}
                        pageSize={listPageSize}
                        total={listTotal}
                        onChange={setListPage}
                        showSizeChanger={false}
                        size="small"
                      />
                    </div>
                  )}
                </>
              ) : (
                <Empty
                  description="暂无数据"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              )}
            </div>
          ) : (
            // ========== 详情视图 ==========
            <div className="detail-view">
              <EventDetailContent
                loading={detailLoading}
                detail={detailData}
                columns={backendColumns}
              />
            </div>
          )}
        </Card>
      </div>
    </Spin>
  );
};

export default EventStatistics;
