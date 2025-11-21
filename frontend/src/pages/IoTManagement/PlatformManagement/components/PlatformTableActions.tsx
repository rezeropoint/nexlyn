import type { PlatformMetadata } from "@/services/iot";
import { Space } from "antd";
import React from "react";
import styles from "./PlatformTableActions.less";

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
      <a onClick={() => onEdit(record)}>编辑</a>
      <a onClick={() => onDelete(record)} className={styles.deleteButton}>
        删除
      </a>
    </Space>
  );
};

export default PlatformTableActions;
