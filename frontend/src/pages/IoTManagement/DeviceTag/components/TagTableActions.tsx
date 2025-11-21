import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import { Button, Space } from "antd";
import React from "react";
import type { DeviceTagSummary } from "../types";

interface TagTableActionsProps {
  record: DeviceTagSummary;
  onEdit: (record: DeviceTagSummary) => void;
  onDelete: (record: DeviceTagSummary) => void;
}

/**
 * 标签表格操作按钮组件
 */
const TagTableActions: React.FC<TagTableActionsProps> = ({
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

export default TagTableActions;
