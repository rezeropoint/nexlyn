/**
 * 信息原子类型表格 - 操作按钮组件
 */

import type { InfoAtomType } from '@/services/lynxmanager';
import { DeleteOutlined, EditOutlined } from '@ant-design/icons';
import { Button, Space, Tooltip } from 'antd';
import React from 'react';

interface InfoAtomTypeTableActionsProps {
  record: InfoAtomType;
  onEdit: (record: InfoAtomType) => void;
  onDelete: (record: InfoAtomType) => void;
}

/**
 * 表格操作按钮组件
 */
const InfoAtomTypeTableActions: React.FC<InfoAtomTypeTableActionsProps> = ({
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

export default InfoAtomTypeTableActions;
