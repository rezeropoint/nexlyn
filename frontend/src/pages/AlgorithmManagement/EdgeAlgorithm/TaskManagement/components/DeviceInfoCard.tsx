import {
  CheckCircleOutlined,
  EnvironmentOutlined,
  ExperimentOutlined,
  MinusCircleOutlined,
} from "@ant-design/icons";
import { Badge, Button, Descriptions, Space, Typography } from "antd";
import React from "react";
import type { AIBoxDeviceOption } from "../types";
import styles from "./DeviceInfoCard.less";

const { Text } = Typography;

interface DeviceInfoCardProps {
  device: AIBoxDeviceOption | null; // 当前选中的设备
  onViewCapabilities?: () => void; // 查看算法能力
}

/**
 * 设备信息卡片
 *
 * 功能：
 * - 展示选中AI Box设备的详细信息
 * - 在线状态、位置、型号等基本信息
 * - 快捷操作按钮
 */
const DeviceInfoCard: React.FC<DeviceInfoCardProps> = ({
  device,
  onViewCapabilities,
}) => {
  if (!device) {
    return (
      <div className={styles.emptyContainer}>
        <Text type="secondary">
          请在左侧选择组织，然后在中间列表中选择AI Box设备
        </Text>
      </div>
    );
  }

  return (
    <div className={styles.deviceContainer}>
      {/* 设备标题和状态 */}
      <div className={styles.deviceHeader}>
        <Space size="middle">
          {device.isOnline ? (
            <Badge status="success" />
          ) : (
            <Badge status="default" />
          )}
          <div>
            <Text strong className={styles.deviceNamePrimary}>
              {device.deviceAlias || device.deviceName}
            </Text>
            {device.deviceAlias && (
              <div>
                <Text type="secondary" className={styles.deviceNameSecondary}>
                  {device.deviceName}
                </Text>
              </div>
            )}
          </div>
        </Space>
        <Button
          type="primary"
          size="small"
          icon={<ExperimentOutlined />}
          onClick={onViewCapabilities}
          disabled={!device.isOnline}
          title={
            device.isOnline
              ? "查看设备支持的算法能力"
              : "设备离线，无法查看算法能力"
          }
        >
          查看算法能力
        </Button>
      </div>

      {/* 设备详细信息 */}
      <Descriptions column={2} size="small" bordered>
        <Descriptions.Item label="设备ID">
          <Text code copyable>
            {device.deviceId}
          </Text>
        </Descriptions.Item>
        <Descriptions.Item label="在线状态">
          {device.isOnline ? (
            <Space>
              <CheckCircleOutlined className={styles.onlineIcon} />
              <Text className={styles.statusOnline}>在线</Text>
            </Space>
          ) : (
            <Space>
              <MinusCircleOutlined className={styles.offlineIcon} />
              <Text type="secondary">离线</Text>
            </Space>
          )}
        </Descriptions.Item>
        <Descriptions.Item label="安装位置" span={2}>
          {device.location ? (
            <Space>
              <EnvironmentOutlined className={styles.locationIcon} />
              <Text>{device.location}</Text>
            </Space>
          ) : (
            <Text type="secondary">未设置</Text>
          )}
        </Descriptions.Item>
      </Descriptions>
    </div>
  );
};

export default DeviceInfoCard;
