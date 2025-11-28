/**
 * 逻辑图配置卡片组件
 */

import type { GraphConfigMetadata } from '@/services/lynxmanager';
import {
  ApiOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CopyOutlined,
  DeleteOutlined,
  EyeOutlined,
  NodeIndexOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Button, Card, Divider, Space, Statistic, Tag, theme, Tooltip } from 'antd';
import React from 'react';
import { getIconColor, getIconComponent } from '../iconConfig';
import styles from './GraphConfigCard.less';

interface GraphConfigCardProps {
  data: GraphConfigMetadata;
  onView: (record: GraphConfigMetadata) => void;
  onCopy: (record: GraphConfigMetadata) => void;
  onDelete: (record: GraphConfigMetadata) => void;
  bodyPadding?: number;
}

/**
 * 单个逻辑图配置卡片
 */
const GraphConfigCard: React.FC<GraphConfigCardProps> = ({
  data,
  onView,
  onCopy,
  onDelete,
  bodyPadding = 16,
}) => {
  const { token } = theme.useToken();

  // 获取图标组件，如果没有设置图标则使用默认图标
  const IconComponent = getIconComponent(data.icon) || getIconComponent('RiFlowChart');
  const iconColor = data.iconColor || getIconColor(data.icon, token) || token.colorPrimary;

  return (
    <Card
      className={styles.graphCard}
      hoverable
      onClick={() => onView(data)}
      bodyStyle={{ padding: bodyPadding }}
    >
      <div className={styles.cardContent}>
        {/* 头部：图标 + 图名称 + 版本 + 状态 */}
        <div className={styles.header}>
          <div className={styles.titleRow}>
            <div className={styles.iconWrapper}>
              {IconComponent && <IconComponent size={40} color={iconColor} />}
            </div>
            <div className={styles.titleContent}>
              <Tooltip title={data.name}>
                <div className={styles.title}>{data.name}</div>
              </Tooltip>
              <Tag color="blue">{data.version}</Tag>
            </div>
          </div>
          <div className={styles.statusRow}>
            <Tag
              icon={data.isEnabled ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
              color={data.isEnabled ? 'success' : 'default'}
            >
              {data.isEnabled ? '已启用' : '未启用'}
            </Tag>
          </div>
        </div>

        <Divider className={styles.divider} />

        {/* 描述信息 */}
        <div className={data.description ? styles.description : styles.emptyDescription}>
          {data.description || '暂无描述'}
        </div>

        {/* 标签 */}
        <div className={styles.tags}>
          {data.tags && data.tags.length > 0 ? (
            // Phase 2.5: 改为使用对象形式（tag.id, tag.name）
            data.tags.map((tag) => (
              <Tag key={tag.id} icon={<ApiOutlined />}>
                {tag.name}
              </Tag>
            ))
          ) : (
            <span className={styles.emptyDescription}>暂无标签</span>
          )}
        </div>

        <Divider className={styles.divider} />

        {/* 运行统计 */}
        <div className={styles.statistics}>
          <Statistic
            title="节点数"
            value={data.nodeCount || 0}
            prefix={<NodeIndexOutlined />}
          />
          <Statistic
            title="执行次数"
            value={data.executionCount || 0}
            prefix={<ThunderboltOutlined />}
          />
          <Statistic
            title="成功率"
            value={data.successRate || 0}
            suffix="%"
            prefix={
              data.successRate && data.successRate >= 80 ? (
                <CheckCircleOutlined className={styles.successIcon} />
              ) : (
                <CloseCircleOutlined className={styles.warningIcon} />
              )
            }
          />
        </div>

        <Divider className={styles.divider} />

        {/* 操作按钮 */}
        <div className={styles.actions} onClick={(e) => e.stopPropagation()}>
          <Space size="small">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => onView(data)}
            >
              查看详情
            </Button>
            <Button
              type="link"
              size="small"
              icon={<CopyOutlined />}
              onClick={() => onCopy(data)}
            >
              复制
            </Button>
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => onDelete(data)}
            >
              删除
            </Button>
          </Space>
        </div>
      </div>
    </Card>
  );
};

export default GraphConfigCard;
