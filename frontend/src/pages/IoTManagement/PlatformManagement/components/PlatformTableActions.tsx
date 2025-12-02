import type { PlatformMetadata } from "@/services/iot";
import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import { Button, Popconfirm, Space } from "antd";
import React from "react";

interface PlatformTableActionsProps {
  record: PlatformMetadata;
  onEdit: (record: PlatformMetadata) => void;
  onDelete: (record: PlatformMetadata) => void;
}

/**
 * 平台配置表格操作按钮组件
 */
const PlatformTableActions: React.FC<PlatformTableActionsProps> = ({
  record,
  onEdit,
  onDelete,
}) => {
  return (
    <Space size="small">
      <Button
        type="link"
        size="small"
        icon={<EditOutlined />}
        onClick={() => onEdit(record)}
      >
        编辑
      </Button>
      <Popconfirm
        title="确认删除"
        description="删除后将无法恢复，确认删除该平台配置？"
        onConfirm={() => onDelete(record)}
        okText="确认"
        cancelText="取消"
      >
        <Button type="link" size="small" danger icon={<DeleteOutlined />}>
          删除
        </Button>
      </Popconfirm>
    </Space>
  );
};

export default PlatformTableActions;
