/**
 * 逻辑图配置表格操作按钮组件
 */

import type { GraphConfigMetadata } from '@/services/lynxmanager';
import { CopyOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import { Button, Space } from 'antd';
import React from 'react';

interface GraphConfigTableActionsProps {
  record: GraphConfigMetadata;
  onView: (record: GraphConfigMetadata) => void;
  onCopy: (record: GraphConfigMetadata) => void;
  onDelete: (record: GraphConfigMetadata) => void;
}

/**
 * 逻辑图配置表格操作按钮
 */
const GraphConfigTableActions: React.FC<GraphConfigTableActionsProps> = ({
  record,
  onView,
  onCopy,
  onDelete,
}) => {
  return (
    <Space size={0}>
      <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => onView(record)}>
        查看详情
      </Button>
      <Button type="link" size="small" icon={<CopyOutlined />} onClick={() => onCopy(record)}>
        复制
      </Button>
      <Button
        type="link"
        size="small"
        danger
        icon={<DeleteOutlined />}
        onClick={() => onDelete(record)}
      >
        删除
      </Button>
    </Space>
  );
};

export default GraphConfigTableActions;
