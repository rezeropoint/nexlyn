/**
 * 工作台仪表盘页面
 * 引入可视化大屏的关键模块，统一工作台视觉体验
 */
import {
  ArrowRightOutlined,
  CarryOutOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DashboardOutlined,
  ReloadOutlined,
  ThunderboltOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import type { UnifiedDeviceStatisticsData } from "@/services/device";
import { getUnifiedDeviceStatistics } from "@/services/device";
import {
  getPendingStats,
  getDurationStats,
  getStatusStats,
  getTrendStats,
  getUserStats,
  getUserAssignments,
} from "@/services/workbench";
import type {
  Assignment,
  GetDurationStatsResponse,
  GetPendingStatsResponse,
  GetStatusStatsResponse,
  GetTrendStatsResponse,
  GetUserStatsResponse,
} from "@/services/workbench/types";
import { useApp } from "@/utils/appContext";
import { Column, Line, Pie } from "@ant-design/plots";
import { Button, Card, Empty, Spin, Tag, Typography, theme, Progress, Pagination } from "antd";
import dayjs from "dayjs";
import "dayjs/locale/zh-cn";
import React, { useEffect, useMemo, useState } from "react";
import { history, useModel } from "umi";
import styles from "./index.less";

const { Text, Title } = Typography;

dayjs.locale("zh-cn");

/**
 * 提取错误消息（兼容 msg 和 message 字段）
 */
const getErrorMessage = (error: any, defaultMsg: string): string => {
  return error?.msg || error?.message || error?.data?.msg || error?.response?.data?.msg || defaultMsg;
};

const Dashboard: React.FC = () => {
  const { token } = theme.useToken();
  const { message } = useApp();
  const { initialState } = useModel("@@initialState");
  const currentUser = initialState?.currentUser;

  const [pendingStats, setPendingStats] = useState<GetPendingStatsResponse["data"]>();
  const [durationStats, setDurationStats] = useState<GetDurationStatsResponse["data"]>();
  const [statusStats, setStatusStats] = useState<GetStatusStatsResponse["data"]>();
  const [trendStats, setTrendStats] = useState<GetTrendStatsResponse["data"]>();
  const [userStats, setUserStats] = useState<GetUserStatsResponse["data"]>();
  const [deviceStats, setDeviceStats] = useState<UnifiedDeviceStatisticsData | null>(null);
  const [priorityTasks, setPriorityTasks] = useState<Assignment[]>([]);
  const [taskTotal, setTaskTotal] = useState(0);
  const [taskPage, setTaskPage] = useState(1);
  const taskPageSize = 7; // 固定每页7条

  const [statsLoading, setStatsLoading] = useState(false);
  const [deviceLoading, setDeviceLoading] = useState(false);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  /**
   * 刷新所有数据
   */
  const handleRefresh = () => {
    setRefreshTrigger((prev) => prev + 1);
    message.success("正在刷新数据...");
  };

  /**
   * 加载事件统计数据（个人所有时间的数据）
   */
  useEffect(() => {
    const orgId = currentUser?.organizationIds?.[0];
    if (!orgId) return;

    const loadStats = async () => {
      setStatsLoading(true);
      try {
        const params = { orgId };

        const [pendingRes, durationRes, statusRes, trendRes, userRes] = await Promise.all([
          getPendingStats(params),
          getDurationStats(params),
          getStatusStats(params),
          getTrendStats({ ...params, groupBy: "day" }),
          getUserStats({ ...params, topN: 8 }),
        ]);

        // 分别处理每个接口响应，允许部分成功
        if (pendingRes.code === 0) {
          setPendingStats(pendingRes.data);
        } else {
          console.error("待处理统计失败:", pendingRes);
          message.warning(`待处理统计加载失败: ${getErrorMessage(pendingRes, "未知错误")}`);
        }

        if (durationRes.code === 0) {
          setDurationStats(durationRes.data);
        } else {
          console.error("时长统计失败:", durationRes);
          message.warning(`时长统计加载失败: ${getErrorMessage(durationRes, "未知错误")}`);
        }

        if (statusRes.code === 0) {
          setStatusStats(statusRes.data);
        } else {
          console.error("状态统计失败:", statusRes);
          message.warning(`状态统计加载失败: ${getErrorMessage(statusRes, "未知错误")}`);
        }

        if (trendRes.code === 0) {
          setTrendStats(trendRes.data);
        } else {
          console.error("趋势统计失败:", trendRes);
          message.warning(`趋势统计加载失败: ${getErrorMessage(trendRes, "未知错误")}`);
        }

        if (userRes.code === 0) {
          setUserStats(userRes.data);
        } else {
          console.error("处理人统计失败:", userRes);
          message.warning(`处理人统计加载失败: ${getErrorMessage(userRes, "未知错误")}`);
        }
      } catch (error: any) {
        console.error("加载事件统计失败:", error);
        message.error(getErrorMessage(error, "加载仪表盘统计失败"));
      } finally {
        setStatsLoading(false);
      }
    };

    loadStats();
  }, [currentUser?.organizationIds, message, refreshTrigger]);

  /**
   * 加载设备统计（个人可见的所有设备）
   */
  useEffect(() => {
    const orgId = currentUser?.organizationIds?.[0];
    if (!orgId) return;

    const loadDevices = async () => {
      setDeviceLoading(true);
      try {
        const response = await getUnifiedDeviceStatistics({ organizationId: orgId });
        // getUnifiedDeviceStatistics 直接返回 data，不是 {code, msg, data} 格式
        setDeviceStats(response);
      } catch (error: any) {
        console.error("加载设备统计失败:", error);
        message.error(getErrorMessage(error, "加载设备统计失败"));
      } finally {
        setDeviceLoading(false);
      }
    };

    loadDevices();
  }, [currentUser?.organizationIds, message, refreshTrigger]);

  /**
   * 加载待办任务（支持分页）
   */
  useEffect(() => {
    const loadPriorityTasks = async () => {
      setTasksLoading(true);
      try {
        const response = await getUserAssignments({
          category: "processed",
          page: taskPage,
          pageSize: taskPageSize,
        });
        if (response.code === 0) {
          setPriorityTasks(response.data?.list || []);
          setTaskTotal(response.data?.total || 0);
        } else {
          message.error(getErrorMessage(response, "加载待办任务失败"));
        }
      } catch (error: any) {
        console.error("加载待办任务失败:", error);
        message.error(getErrorMessage(error, "加载待办任务失败"));
      } finally {
        setTasksLoading(false);
      }
    };

    loadPriorityTasks();
  }, [taskPage, taskPageSize, message, refreshTrigger]);

  const trendChartData = useMemo(() => {
    if (!trendStats?.timePoints) return [];
    // 转换为双线图数据格式
    const data: Array<{ date: string; type: string; count: number }> = [];
    trendStats.timePoints.forEach((point) => {
      const dateStr = dayjs(point.date).format("MM-DD");
      data.push(
        { date: dateStr, type: "新增事件", count: point.totalCount },
        { date: dateStr, type: "已完成", count: point.completedCount }
      );
    });
    return data;
  }, [trendStats?.timePoints]);

  const statusPieData = useMemo(
    () =>
      (statusStats?.statusCounts || []).map((item) => ({
        type: item.status,
        value: item.count,
      })),
    [statusStats?.statusCounts]
  );

  const deviceHealth = useMemo(() => {
    if (!deviceStats) return { total: 0, online: 0, offline: 0, rate: 0 };
    const total = deviceStats.deviceTotal;
    const online = deviceStats.deviceOnline;
    const rate = total > 0 ? Math.round((online / total) * 100) : 0;
    return { total, online, offline: deviceStats.deviceOffline, rate };
  }, [deviceStats]);

  const trendLineConfig = {
    data: trendChartData,
    xField: "date",
    yField: "count",
    seriesField: "type",
    smooth: true,
    legend: {
      position: "top" as const,
    },
    color: [token.colorPrimary, token.colorSuccess],
    lineStyle: {
      lineWidth: 2,
    },
    point: {
      size: 4,
      shape: "circle",
    },
    tooltip: {
      shared: true,
      showCrosshairs: true,
    },
    padding: "auto" as const,
  };

  const statusPieConfig = {
    data: statusPieData,
    angleField: "value",
    colorField: "type",
    legend: { position: "right" as const },
    innerRadius: 0.6,
    statistic: {
      title: { content: "总计" },
      content: { content: String(statusStats?.total || 0) },
    },
    label: {
      text: (datum: any) => `${datum.type}\n${datum.value}`,
      position: "outside",
      style: {
        fill: "var(--ant-color-text)",
      },
    },
  };

  const userColumnConfig = {
    data: userStats?.userMetrics || [],
    xField: "userName",
    yField: "count",
    colorField: "userName",
    legend: false,
    columnStyle: { radius: [6, 6, 0, 0] },
    tooltip: {
      formatter: (datum: { userName: string; count: number }) => ({
        name: datum.userName,
        value: `${datum.count} 次处理`,
      }),
    },
  };

  const priorityEmpty = !tasksLoading && priorityTasks.length === 0;

  /**
   * 跳转到待办任务页面并打开详情
   */
  const handleViewDetail = (task: Assignment) => {
    if (!task.flowId || !task.journeyId || !task.id) {
      message.warning("流程信息不完整，无法查看详情");
      return;
    }
    // 跳转到待办任务页面，通过 URL 参数传递任务信息（包含assignmentId）
    history.push(`/workbench/pending-tasks?flowId=${task.flowId}&journeyId=${task.journeyId}&assignmentId=${task.id}`);
  };

  const loadingHero = statsLoading;

  const heroDateText = dayjs().format("YYYY年MM月DD日 dddd");
  const heroUpdateText = dayjs().format("HH:mm 更新");

  // 只保留3个最关键KPI（按流程顺序：待处理 → 处理中 → 已完成）
  const heroMetrics = [
    {
      label: "待处理",
      value: pendingStats?.totalPending ?? 0,
      icon: <ClockCircleOutlined />,
      className: styles.metricWarning,
    },
    {
      label: "处理中",
      value: statusStats?.statusCounts?.find((item) => item.statusKey === "processing")?.count ?? 0,
      icon: <CarryOutOutlined />,
      className: styles.metricPrimary,
    },
    {
      label: "已完成",
      value:
        statusStats?.statusCounts?.find((item) => item.statusKey === "finished")?.count ??
        durationStats?.completedCount ??
        0,
      icon: <CheckCircleOutlined />,
      className: styles.metricSuccess,
    },
  ];

  return (
    <PageContainer className={styles.dashboardPage} pageHeaderRender={false}>
      <div className={styles.heroSection}>
        <div className={styles.heroHeader}>
          <div>
            <span className={styles.heroEyebrow}>我的工作台</span>
            <Title level={3} className={styles.heroTitle}>
              <DashboardOutlined /> 审批进度一屏尽览
            </Title>
            <div className={styles.heroDateRow}>
              <Text className={styles.heroDate}>{heroDateText}</Text>
              <Tag className={styles.heroUpdate} bordered={false}>
                {heroUpdateText}
              </Tag>
            </div>
            <Text className={styles.heroDescription}>
              个人工作台 · 掌握关键数据，快速定位待办任务
            </Text>
          </div>
          <div>
            <Button
              type="primary"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={statsLoading || deviceLoading || tasksLoading}
              size="large"
              className={styles.heroRefreshButton}
            >
              刷新数据
            </Button>
          </div>
        </div>

        <Spin spinning={loadingHero}>
          <div className={styles.heroMetricsGrid}>
            {heroMetrics.map((metric) => (
              <div key={metric.label} className={`${styles.metricCard} ${metric.className}`}>
                <span className={styles.metricIcon}>{metric.icon}</span>
                <div className={styles.metricValue}>{metric.value}</div>
                <div className={styles.metricLabel}>{metric.label}</div>
              </div>
            ))}
          </div>
        </Spin>
      </div>

      <div className={styles.mainLayout}>
        <div className={styles.leftColumn}>
          {/* 事件健康度 + 设备运行概览并排 */}
          <div className={styles.healthRow}>
            <Card
              className={styles.sectionCard}
              title={
                <span>
                  <ThunderboltOutlined style={{ marginRight: 8, color: "var(--ant-color-primary)" }} />
                  流程状态
                </span>
              }
              extra={<span className={styles.sectionHint}>实时统计</span>}
            >
              <Spin spinning={statsLoading}>
                {statusPieData.length > 0 ? (
                  <div className={styles.eventHealth}>
                    <div className={styles.eventChart}>
                      <Pie {...statusPieConfig} height={240} />
                    </div>
                    <div className={styles.eventList}>
                      {statusStats?.statusCounts?.map((item) => (
                        <div key={item.statusKey} className={styles.eventItem}>
                          <div className={styles.eventItemLabel}>{item.status}</div>
                          <div className={styles.eventItemValue}>{item.count}</div>
                          <Tag className={styles.eventTag}>{item.percentage.toFixed(1)}%</Tag>
                        </div>
                      ))}
                    </div>
                  </div>
                ) : (
                  <Empty description="暂无事件统计" />
                )}
              </Spin>
            </Card>

            <Card className={styles.sectionCard} title="设备运行概览" extra={<span className={styles.sectionHint}>可视化大屏精简版</span>}>
              <Spin spinning={deviceLoading}>
                {deviceStats ? (
                  <div className={styles.deviceOverview}>
                    <div className={styles.deviceMetricItem}>
                      <div className={styles.deviceValue}>{deviceHealth.total}</div>
                      <div className={styles.deviceLabel}>总设备</div>
                    </div>
                    <div className={styles.deviceMetricItem}>
                      <div className={styles.deviceValue}>{deviceHealth.online}</div>
                      <div className={styles.deviceLabel}>在线数</div>
                    </div>
                    <div className={styles.deviceMetricItem}>
                      <div className={styles.deviceValue}>{deviceStats.channelTotal ?? 0}</div>
                      <div className={styles.deviceLabel}>监控点</div>
                    </div>
                    <div className={styles.deviceProgress}>
                      <div className={styles.progressHeader}>
                        <span>在线率</span>
                        <span className={styles.deviceRate}>{deviceHealth.rate}%</span>
                      </div>
                      <Progress
                        percent={deviceHealth.rate}
                        strokeColor={token.colorSuccess}
                        trailColor="var(--ant-color-fill-tertiary)"
                        showInfo={false}
                        size={["100%", 8]}
                      />
                    </div>
                  </div>
                ) : (
                  <Empty description="暂无设备数据" />
                )}
              </Spin>
            </Card>
          </div>

          <div className={styles.analyticsRow}>
            <Card className={styles.sectionCard} title="事件趋势对比" extra={<span className={styles.sectionHint}>新增 vs 完成</span>}>
              <Spin spinning={statsLoading}>
                {trendChartData.length > 0 ? (
                  <Line {...trendLineConfig} height={200} />
                ) : (
                  <Empty description="暂无趋势数据" />
                )}
              </Spin>
            </Card>
            <Card className={styles.sectionCard} title="处理人效率 TOP8">
              <Spin spinning={statsLoading}>
                {userStats?.userMetrics && userStats.userMetrics.length > 0 ? (
                  <Column {...userColumnConfig} height={200} />
                ) : (
                  <Empty description="暂无处理人数据" />
                )}
              </Spin>
            </Card>
          </div>
        </div>

        <div className={styles.rightColumn}>
          <Card
            className={styles.priorityCard}
            title={
              <span>
                <ClockCircleOutlined style={{ marginRight: 8, color: "var(--ant-color-warning)" }} />
                待办任务
              </span>
            }
            extra={
              <Button
                type="text"
                icon={<ArrowRightOutlined />}
                onClick={() => history.push("/workbench/pending-tasks")}
              >
                查看全部
              </Button>
            }
          >
            <Spin spinning={tasksLoading}>
              <div className={styles.taskList}>
                {priorityEmpty ? (
                  <Empty description="暂无待办任务" />
                ) : (
                  priorityTasks.map((task) => (
                    <div key={task.id} className={styles.taskItem}>
                      <div className={styles.taskContent}>
                        <div className={styles.taskTitle}>{task.flowTitle || `流程 #${task.flowId}`}</div>
                        <div className={styles.taskMeta}>
                          <Tag bordered={false} className={styles.taskTag}>
                            {task.status === "processing" ? "处理中" : "已完成"}
                          </Tag>
                          <span>{dayjs(task.updatedAt).format("MM-DD HH:mm")}</span>
                        </div>
                      </div>
                      <Button
                        type="primary"
                        size="small"
                        onClick={() => handleViewDetail(task)}
                        className={styles.taskAction}
                      >
                        处理
                      </Button>
                    </div>
                  ))
                )}
              </div>
              {!priorityEmpty && (
                <div className={styles.taskPagination}>
                  <Pagination
                    current={taskPage}
                    pageSize={taskPageSize}
                    total={taskTotal}
                    onChange={(page) => {
                      setTaskPage(page);
                    }}
                    showQuickJumper
                    showTotal={(total) => `共 ${total} 条`}
                    size="small"
                  />
                </div>
              )}
            </Spin>
          </Card>
        </div>
      </div>

    </PageContainer>
  );
};

export default Dashboard;
