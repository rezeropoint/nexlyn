/**
 * 标签表格 - 操作按钮组件
 */

import type { LynxTag } from '@/services/lynxmanager';
import { DeleteOutlined, EditOutlined } from '@ant-design/icons';
import { Button, Space, Tooltip } from 'antd';
import React from 'react';

interface TagTableActionsProps {
  record: LynxTag;
  onEdit: (record: LynxTag) => void;
  onDelete: (record: LynxTag) => void;
}

/**
 * 表格操作按钮组件
 */
const TagTableActions: React.FC<TagTableActionsProps> = ({
  record,
  onEdit,
  onDelete,
}) => {
  return (
    <Space size="small">
      <Tooltip title="编辑">
        <Button
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => onEdit(record)}
        >
          编辑
        </Button>
      </Tooltip>
      <Tooltip title="删除">
        <Button
          type="link"
          size="small"
          danger
          icon={<DeleteOutlined />}
          onClick={() => onDelete(record)}
        >
          删除
        </Button>
      </Tooltip>
    </Space>
  );
};

export default TagTableActions;
