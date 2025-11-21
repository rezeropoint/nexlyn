import {
  DisconnectOutlined,
  EditOutlined,
  EyeOutlined,
  MoreOutlined,
} from "@ant-design/icons";
import type { MenuProps } from "antd";
import { Button, Dropdown, Space } from "antd";
import React from "react";
import type { DeviceBindingSummary } from "../types";

interface DeviceTableActionsProps {
  record: DeviceBindingSummary;
  onView: (record: DeviceBindingSummary) => void;
  onEdit: (record: DeviceBindingSummary) => void;
  onUnbind: (record: DeviceBindingSummary) => void;
}

/**
 * 设备表格操作按钮组件
 */
const DeviceTableActions: React.FC<DeviceTableActionsProps> = ({
  record,
  onView,
  onEdit,
  onUnbind,
}) => {
  const menuItems: MenuProps["items"] = [
    {
      key: "edit",
      icon: <EditOutlined />,
      label: "编辑设备",
      onClick: () => onEdit(record),
    },
    {
      type: "divider",
    },
    {
      key: "unbind",
      icon: <DisconnectOutlined />,
      label: "解绑设备",
      danger: true,
      onClick: () => onUnbind(record),
    },
  ];

  return (
    <Space size="small">
      <Button
        type="link"
        size="small"
        icon={<EyeOutlined />}
        onClick={() => onView(record)}
      >
        详情
      </Button>
      <Dropdown
        menu={{ items: menuItems }}
        trigger={["click"]}
        placement="bottomRight"
      >
        <Button type="link" size="small" icon={<MoreOutlined />}>
          操作
        </Button>
      </Dropdown>
    </Space>
  );
};

export default DeviceTableActions;
