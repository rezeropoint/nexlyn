import { formatDateTime } from "@/utils/date";
import { Badge, Descriptions, Drawer, Space, Tag } from "antd";
import React from "react";
import { DEVICE_STATUS_MAP } from "../constants";
import type { DeviceBinding } from "../types";

interface DeviceDetailDrawerProps {
  open: boolean;
  onClose: () => void;
  device?: DeviceBinding;
}

/**
 * 设备详情抽屉组件
 */
const DeviceDetailDrawer: React.FC<DeviceDetailDrawerProps> = ({
  open,
  onClose,
  device,
}) => {
  if (!device) return null;

  return (
    <Drawer
      title="设备详情"
      width={720}
      open={open}
      onClose={onClose}
      destroyOnHidden
    >
      <Descriptions column={2} bordered>
        <Descriptions.Item label="设备ID" span={2}>
          {device.deviceId}
        </Descriptions.Item>

        <Descriptions.Item label="设备名称">
          {device.deviceName}
        </Descriptions.Item>

        <Descriptions.Item label="设备别名">
          {device.deviceAlias || "-"}
        </Descriptions.Item>

        <Descriptions.Item label="设备型号">
          {device.deviceModel}
        </Descriptions.Item>

        <Descriptions.Item label="设备类别">
          {device.deviceCategory}
        </Descriptions.Item>

        <Descriptions.Item label="在线状态">
          {device.isOnline ? (
            <Badge status="success" text="在线" />
          ) : (
            <Badge status="default" text="离线" />
          )}
        </Descriptions.Item>

        <Descriptions.Item label="设备状态">
          <Tag
            color={
              DEVICE_STATUS_MAP[device.status as keyof typeof DEVICE_STATUS_MAP]
                ?.color
            }
          >
            {DEVICE_STATUS_MAP[device.status as keyof typeof DEVICE_STATUS_MAP]
              ?.text || device.status}
          </Tag>
        </Descriptions.Item>

        <Descriptions.Item label="所属组织" span={2}>
          {device.orgId}
        </Descriptions.Item>

        <Descriptions.Item label="安装位置" span={2}>
          {device.location || "-"}
        </Descriptions.Item>

        <Descriptions.Item label="安装日期" span={2}>
          {device.installationDate || "-"}
        </Descriptions.Item>

        <Descriptions.Item label="最后数据时间" span={2}>
          {device.lastDataAt ? formatDateTime(device.lastDataAt) : "-"}
        </Descriptions.Item>

        <Descriptions.Item label="设备标签" span={2}>
          {device.tags && device.tags.length > 0 ? (
            <Space wrap>
              {device.tags.map((tag) => (
                <Tag key={tag.id} color={tag.color}>
                  {tag.name}
                </Tag>
              ))}
            </Space>
          ) : (
            "-"
          )}
        </Descriptions.Item>

        <Descriptions.Item label="设备描述" span={2}>
          {device.description || "-"}
        </Descriptions.Item>

        <Descriptions.Item label="创建时间">
          {formatDateTime(device.createdAt)}
        </Descriptions.Item>

        <Descriptions.Item label="更新时间">
          {formatDateTime(device.updatedAt)}
        </Descriptions.Item>
      </Descriptions>
    </Drawer>
  );
};

export default DeviceDetailDrawer;
