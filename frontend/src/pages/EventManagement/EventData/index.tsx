import useResponsiveLayout from "@/hooks/useResponsiveLayout";
import { getEventConfigList, getStatusStats } from "@/services/eventhandler";
import type { EventConfigWithFields } from "@/pages/EventManagement/types";
import {
  ArrowLeftOutlined,
  CompressOutlined,
  ExpandAltOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import {
  PageContainer,
  ProCard,
} from "@ant-design/pro-components";
import { Button, Space, Splitter, Tag } from "antd";
import { useEffect, useRef, useState } from "react";
import EventDataTable from "./components/EventDataTable";
import EventConfigCardSelector from "./components/EventConfigCardSelector";
import OrganizationSelector, {
  type OrganizationSelectorRef,
} from "./components/OrganizationSelector";
import styles from "./index.less";

/**
 * 事件数据查询页面
 * 左侧：组织树选择器（可拖拽调整宽度）
 * 右侧：事件选择器 + 状态筛选 + 数据表格
 */
const EventData: React.FC = () => {
  const orgSelectorRef = useRef<OrganizationSelectorRef>(null);
  const [selectedOrgId, setSelectedOrgId] = useState<string>("");
  const [selectedEventConfig, setSelectedEventConfig] = useState<EventConfigWithFields | null>(null);
  const [orgTreeLoading, setOrgTreeLoading] = useState(false);
  const [eventConfigs, setEventConfigs] = useState<EventConfigWithFields[]>([]);
  const [eventConfigsLoading, setEventConfigsLoading] = useState(false);

  // 事件统计数据
  const [eventSummary, setEventSummary] = useState({
    total: 0,
    pending: 0,
    processing: 0,
    finished: 0,
  });

  // 响应式布局：屏幕宽度 < 1600px 时自动切换为垂直布局
  const layout = useResponsiveLayout({ breakpoint: "xl" });

  // 加载事件配置列表
  useEffect(() => {
    const loadEventConfigs = async () => {
      setEventConfigsLoading(true);
      try {
        const response = await getEventConfigList({
          pageSize: 1000,
          enabled: true,
        });
        if (response.code === 0 && response.data?.list) {
          setEventConfigs(response.data.list);
        }
      } catch (error) {
        console.error("加载事件配置列表失败:", error);
      } finally {
        setEventConfigsLoading(false);
      }
    };

    loadEventConfigs();
  }, []);

  // 处理组织选择
  const handleOrgSelect = (orgId: string) => {
    setSelectedOrgId(orgId);
  };

  // 加载统计数据
  useEffect(() => {
    const loadStats = async () => {
      // 只有在选择了事件配置和组织时才加载统计
      if (!selectedEventConfig?.id || !selectedOrgId) {
        return;
      }

      try {
        const response = await getStatusStats({
          eventIds: [selectedEventConfig.id],
          orgId: selectedOrgId,
        });

        if (response.code === 0 && response.data) {
          const { statusCounts, total } = response.data;

          // 从状态计数数组中提取特定状态的数量
          const pendingCount =
            statusCounts.find((item) => item.statusKey === "pending")?.count || 0;
          const processingCount =
            statusCounts.find((item) => item.statusKey === "processing")?.count || 0;
          const finishedCount =
            statusCounts.find((item) => item.statusKey === "finished")?.count || 0;

          setEventSummary({
            total,
            pending: pendingCount,
            processing: processingCount,
            finished: finishedCount,
          });
        }
      } catch (error) {
        // 统计接口失败时优雅降级，不影响主要功能
        console.error("加载统计数据失败:", error);
        // 重置为0，避免显示旧数据
        setEventSummary({ total: 0, pending: 0, processing: 0, finished: 0 });
      }
    };

    loadStats();
  }, [selectedEventConfig?.id, selectedOrgId]);

  // 组织树工具栏
  const orgToolbar = (
    <Space size={4}>
      <Button
        size="small"
        icon={<ExpandAltOutlined />}
        onClick={() => orgSelectorRef.current?.expandAll()}
        title="全部展开"
      />
      <Button
        size="small"
        icon={<CompressOutlined />}
        onClick={() => orgSelectorRef.current?.collapseAll()}
        title="全部收起"
      />
      <Button
        size="small"
        icon={<ReloadOutlined />}
        onClick={() => {
          setOrgTreeLoading(true);
          orgSelectorRef.current?.refresh();
          setTimeout(() => setOrgTreeLoading(false), 300);
        }}
        loading={orgTreeLoading}
        title="刷新"
      />
    </Space>
  );

  return (
    <PageContainer
      header={{
        title: "事件数据查询",
        subTitle: "查询和分析Skylark流程事件数据",
        extra:
          selectedEventConfig?.id && selectedOrgId
            ? [
                <Space key="stats">
                  <Tag color="blue">总数: {eventSummary.total}</Tag>
                  <Tag color="orange">待处理: {eventSummary.pending}</Tag>
                  <Tag color="processing">处理中: {eventSummary.processing}</Tag>
                  <Tag color="success">已完成: {eventSummary.finished}</Tag>
                </Space>,
              ]
            : undefined,
      }}
    >
      {/* 左右分栏布局（支持拖拽调整、响应式切换） */}
      <Splitter layout={layout} className={styles.splitterContainer}>
        {/* 左侧组织树面板 */}
        <Splitter.Panel
          defaultSize="25%"
          min={180}
          max="45%"
          collapsible={{ start: true }}
          className={styles.leftPanel}
        >
          <ProCard title="组织架构" extra={orgToolbar}>
            <OrganizationSelector
              ref={orgSelectorRef}
              selectedOrgId={selectedOrgId}
              onSelect={handleOrgSelect}
              hideToolbar={true}
            />
          </ProCard>
        </Splitter.Panel>

        {/* 右侧内容区面板 */}
        <Splitter.Panel min={600} max="80%" className={styles.rightPanel}>
          {/* 未选择事件时：显示事件配置卡片选择器 */}
          {!selectedEventConfig ? (
            <ProCard>
              <div className={styles.eventConfigSelector}>
                <EventConfigCardSelector
                  eventConfigs={eventConfigs}
                  value={undefined}
                  onChange={setSelectedEventConfig}
                  loading={eventConfigsLoading}
                  selectedOrgId={selectedOrgId}
                />
              </div>
            </ProCard>
          ) : (
            /* 选择事件后：显示数据表格区域 */
            <ProCard
              title={`事件：${selectedEventConfig.name}`}
              extra={
                <Button
                  type="link"
                  icon={<ArrowLeftOutlined />}
                  onClick={() => setSelectedEventConfig(null)}
                >
                  重新选择事件
                </Button>
              }
            >
              {/* 数据表格 */}
              {selectedOrgId ? (
                <EventDataTable
                  eventConfig={selectedEventConfig}
                  selectedOrgId={selectedOrgId}
                />
              ) : (
                <div className={styles.emptyState}>请先选择组织</div>
              )}
            </ProCard>
          )}
        </Splitter.Panel>
      </Splitter>
    </PageContainer>
  );
};

export default EventData;
