/**
 * 事件配置卡片选择器
 * 以卡片形式展示事件配置列表，支持搜索过滤和统计数据展示
 */
import {
  CheckCircleFilled,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ColumnHeightOutlined,
  FileTextOutlined,
  LoadingOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { EventConfigWithFields } from "@/pages/EventManagement/types";
import { getStatusStats } from "@/services/eventhandler";
import { Button, Card, Col, Dropdown, Empty, Flex, Input, Row, Skeleton, Space, Tag, Tooltip, Typography } from "antd";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import type { MenuProps } from "antd";
import styles from "./EventConfigCardSelector.less";

const { Title, Paragraph } = Typography;
const { Search } = Input;

/** 卡片密度类型 */
type CardDensity = "loose" | "default" | "compact";

export interface EventConfigCardSelectorProps {
  /** 事件配置列表 */
  eventConfigs: EventConfigWithFields[];
  /** 当前选中的事件配置ID */
  value?: string;
  /** 选中事件配置时的回调 */
  onChange?: (config: EventConfigWithFields) => void;
  /** 是否加载中 */
  loading?: boolean;
  /** 当前选中的组织ID（用于加载统计数据） */
  selectedOrgId?: string;
}

/** 单个事件配置的统计数据（基于 getStatusStats 接口） */
interface CardEventStats {
  eventConfigId: string;
  total: number;      // 总数
  pending: number;    // 待处理
  processing: number; // 处理中
  finished: number;   // 已完成
}

/**
 * 事件配置卡片选择器组件
 */
const EventConfigCardSelector: React.FC<EventConfigCardSelectorProps> = ({
  eventConfigs,
  value,
  onChange,
  loading: _loading = false,
  selectedOrgId,
}) => {
  const [searchKeyword, setSearchKeyword] = useState<string>("");
  const [allEventStats, setAllEventStats] = useState<CardEventStats[]>([]);
  const [statsLoading, setStatsLoading] = useState(false);
  const [statsError, setStatsError] = useState<string>("");
  const [cardDensity, setCardDensity] = useState<CardDensity>("default"); // 密度控制

  // 过滤事件配置列表（根据搜索关键词）
  const filteredConfigs = useMemo(() => {
    if (!searchKeyword.trim()) {
      return eventConfigs;
    }
    const keyword = searchKeyword.trim().toLowerCase();
    return eventConfigs.filter((config) => {
      const matchName = config.name?.toLowerCase().includes(keyword);
      const matchFlowTitle = config.flowTitle?.toLowerCase().includes(keyword);
      const matchDescription = config.description?.toLowerCase().includes(keyword);
      return matchName || matchFlowTitle || matchDescription;
    });
  }, [eventConfigs, searchKeyword]);

  // 加载所有事件配置的统计数据（并发调用 getStatusStats）
  const loadAllStats = useCallback(async () => {
    // 如果没有组织ID，则不加载统计数据
    if (!selectedOrgId) {
      setAllEventStats([]);
      setStatsError("");
      return;
    }

    // 如果没有事件配置，则不加载
    if (eventConfigs.length === 0) {
      setAllEventStats([]);
      return;
    }

    setStatsLoading(true);
    setStatsError("");

    try {
      // 为每个事件配置并发调用 getStatusStats 接口
      const statsPromises = eventConfigs.map(async (config) => {
        try {
          const response = await getStatusStats({
            eventIds: [config.id],
            orgId: selectedOrgId,
          });

          if (response.code === 0 && response.data) {
            const { statusCounts, total } = response.data;

            // 提取各状态的数量
            const pending = statusCounts?.find((s) => s.statusKey === "pending")?.count || 0;
            const processing = statusCounts?.find((s) => s.statusKey === "processing")?.count || 0;
            const finished = statusCounts?.find((s) => s.statusKey === "finished")?.count || 0;

            return {
              eventConfigId: config.id,
              total: total || 0,
              pending,
              processing,
              finished,
            };
          }

          // 接口返回失败时返回默认值
          return {
            eventConfigId: config.id,
            total: 0,
            pending: 0,
            processing: 0,
            finished: 0,
          };
        } catch (error) {
          // 单个请求失败时返回默认值，不影响其他请求
          console.error(`加载事件配置 ${config.id} 的统计数据失败:`, error);
          return {
            eventConfigId: config.id,
            total: 0,
            pending: 0,
            processing: 0,
            finished: 0,
          };
        }
      });

      // 等待所有请求完成
      const results = await Promise.all(statsPromises);
      setAllEventStats(results);
      setStatsError(""); // 清除错误状态
    } catch (error) {
      console.error("加载统计数据失败:", error);
      setStatsError("加载统计数据失败");
      setAllEventStats([]);
    } finally {
      setStatsLoading(false);
    }
  }, [selectedOrgId, eventConfigs]);

  // 当组织ID变化时，重新加载统计数据
  useEffect(() => {
    loadAllStats();
  }, [loadAllStats]);

  // 处理卡片选择
  const handleSelectCard = useCallback(
    (config: EventConfigWithFields) => {
      onChange?.(config);
    },
    [onChange]
  );

  // 渲染统计数据标签
  const renderStatsTags = (configId: string) => {
    // 如果没有选择组织，显示提示信息
    if (!selectedOrgId) {
      return (
        <div className={styles.statsPlaceholder}>
          请先选择组织以查看统计
        </div>
      );
    }

    // 如果正在加载，显示骨架屏
    if (statsLoading) {
      return (
        <Row gutter={[8, 8]} className={styles.statsGrid}>
          <Col span={12}>
            <div className={styles.statItemSkeleton}>
              <Skeleton.Button active size="small" style={{ width: 18, height: 18, minWidth: 18 }} />
              <div style={{ flex: 1, marginLeft: 8 }}>
                <Skeleton.Input active size="small" style={{ width: '100%', height: 16, minWidth: 40 }} />
                <Skeleton.Input active size="small" style={{ width: '60%', height: 12, minWidth: 30, marginTop: 4 }} />
              </div>
            </div>
          </Col>
          <Col span={12}>
            <div className={styles.statItemSkeleton}>
              <Skeleton.Button active size="small" style={{ width: 18, height: 18, minWidth: 18 }} />
              <div style={{ flex: 1, marginLeft: 8 }}>
                <Skeleton.Input active size="small" style={{ width: '100%', height: 16, minWidth: 40 }} />
                <Skeleton.Input active size="small" style={{ width: '60%', height: 12, minWidth: 30, marginTop: 4 }} />
              </div>
            </div>
          </Col>
          <Col span={12}>
            <div className={styles.statItemSkeleton}>
              <Skeleton.Button active size="small" style={{ width: 18, height: 18, minWidth: 18 }} />
              <div style={{ flex: 1, marginLeft: 8 }}>
                <Skeleton.Input active size="small" style={{ width: '100%', height: 16, minWidth: 40 }} />
                <Skeleton.Input active size="small" style={{ width: '60%', height: 12, minWidth: 30, marginTop: 4 }} />
              </div>
            </div>
          </Col>
          <Col span={12}>
            <div className={styles.statItemSkeleton}>
              <Skeleton.Button active size="small" style={{ width: 18, height: 18, minWidth: 18 }} />
              <div style={{ flex: 1, marginLeft: 8 }}>
                <Skeleton.Input active size="small" style={{ width: '100%', height: 16, minWidth: 40 }} />
                <Skeleton.Input active size="small" style={{ width: '60%', height: 12, minWidth: 30, marginTop: 4 }} />
              </div>
            </div>
          </Col>
        </Row>
      );
    }

    // 如果加载失败
    if (statsError) {
      return <div className={styles.statsPlaceholder}>{statsError}</div>;
    }

    // 查找对应事件配置的统计数据
    const eventStat = allEventStats.find((stat) => stat.eventConfigId === configId);

    if (!eventStat) {
      return <div className={styles.statsPlaceholder}>暂无数据</div>;
    }

    return (
      <Row gutter={[8, 8]} className={styles.statsGrid}>
        <Col span={12}>
          <div className={styles.statItem}>
            <FileTextOutlined className={styles.statIcon} style={{ color: 'var(--ant-color-primary)' }} />
            <div className={styles.statContent}>
              <div className={styles.statValue}>{eventStat.total}</div>
              <div className={styles.statLabel}>总数</div>
            </div>
          </div>
        </Col>
        <Col span={12}>
          <div className={styles.statItem}>
            <ClockCircleOutlined className={styles.statIcon} style={{ color: 'var(--ant-color-warning)' }} />
            <div className={styles.statContent}>
              <div className={styles.statValue}>{eventStat.pending}</div>
              <div className={styles.statLabel}>待处理</div>
            </div>
          </div>
        </Col>
        <Col span={12}>
          <div className={styles.statItem}>
            <LoadingOutlined className={styles.statIcon} style={{ color: 'var(--ant-color-link)' }} />
            <div className={styles.statContent}>
              <div className={styles.statValue}>{eventStat.processing}</div>
              <div className={styles.statLabel}>处理中</div>
            </div>
          </div>
        </Col>
        <Col span={12}>
          <div className={styles.statItem}>
            <CheckCircleOutlined className={styles.statIcon} style={{ color: 'var(--ant-color-success)' }} />
            <div className={styles.statContent}>
              <div className={styles.statValue}>{eventStat.finished}</div>
              <div className={styles.statLabel}>已完成</div>
            </div>
          </div>
        </Col>
      </Row>
    );
  };

  // 渲染单个事件配置卡片
  const renderCard = (config: EventConfigWithFields) => {
    const isSelected = value === config.id;

    return (
      <div
        key={config.id}
        className={styles.cardWrapper}
      >
        <Card
          hoverable
          className={`${styles.eventCard} ${styles[`density-${cardDensity}`]} ${isSelected ? styles.selected : ""}`}
        >
          {/* 卡片头部：事件名称 + 状态标签 */}
          <div className={styles.cardHeader}>
            <Title level={5} className={styles.eventTitle}>
              <FileTextOutlined style={{ marginRight: 8 }} />
              {config.name}
            </Title>
            <Tag color={config.enabled ? "success" : "default"}>
              {config.enabled ? "已启用" : "已禁用"}
            </Tag>
          </div>

          {/* 流程信息 */}
          {config.flowTitle && (
            <div className={styles.flowInfo}>
              流程: {config.flowTitle} (ID: {config.flowId})
            </div>
          )}

          {/* 描述信息 */}
          {config.description && (
            <Paragraph className={styles.description} ellipsis={{ rows: 2 }}>
              {config.description}
            </Paragraph>
          )}

          {/* 统计数据区域 */}
          <div className={styles.statsSection}>{renderStatsTags(config.id)}</div>

          {/* 选择按钮 */}
          <div className={styles.selectButtonContainer}>
            <Button
              type={isSelected ? "default" : "primary"}
              icon={isSelected ? <CheckCircleFilled /> : undefined}
              onClick={() => handleSelectCard(config)}
              block
            >
              {isSelected ? "已选择" : "选择"}
            </Button>
          </div>
        </Card>
      </div>
    );
  };

  // 密度选择菜单
  const densityMenuItems: MenuProps['items'] = [
    { key: 'loose', label: '宽松', onClick: () => setCardDensity('loose') },
    { key: 'default', label: '中等', onClick: () => setCardDensity('default') },
    { key: 'compact', label: '紧凑', onClick: () => setCardDensity('compact') },
  ];

  return (
    <div className={styles.cardSelectorContainer}>
      {/* 工具栏：搜索框 + 操作按钮 */}
      <div className={styles.toolbar}>
        <div className={styles.searchBox}>
          <Search
            placeholder="搜索事件名称、流程名称或描述..."
            allowClear
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            style={{ maxWidth: 500 }}
          />
        </div>
        <Space size="small">
          <Tooltip title="刷新">
            <Button
              icon={<ReloadOutlined />}
              onClick={loadAllStats}
              loading={statsLoading}
            />
          </Tooltip>
          <Tooltip title="密度">
            <Dropdown menu={{ items: densityMenuItems, selectedKeys: [cardDensity] }}>
              <Button icon={<ColumnHeightOutlined />} />
            </Dropdown>
          </Tooltip>
        </Space>
      </div>

      {/* 卡片展示区域 */}
      <div className={styles.cardsContainer}>
        {filteredConfigs.length > 0 ? (
          <Flex gap={16} wrap="wrap">
            {filteredConfigs.map((config) => renderCard(config))}
          </Flex>
        ) : (
          <Empty
            className={styles.emptyState}
            description={
              searchKeyword.trim()
                ? `未找到匹配 "${searchKeyword}" 的事件配置`
                : "暂无事件配置"
            }
          />
        )}
      </div>
    </div>
  );
};

export default EventConfigCardSelector;
