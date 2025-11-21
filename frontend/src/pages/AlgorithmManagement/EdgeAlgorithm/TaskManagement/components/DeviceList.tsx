import {
  CheckCircleOutlined,
  EnvironmentOutlined,
  MinusCircleOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import { Empty, Input, List, Space, Spin, Typography } from "antd";
import classNames from "classnames";
import React from "react";
import type { AIBoxDeviceOption } from "../types";
import styles from "./DeviceList.less";

const { Text } = Typography;

interface DeviceListProps {
  devices: AIBoxDeviceOption[]; // 设备列表
  selectedDeviceId?: string; // 当前选中的设备ID
  loading?: boolean; // 加载状态
  onSelect?: (deviceId: string) => void; // 设备选择回调
}

/**
 * AI Box 设备列表组件
 *
 * 功能：
 * - 显示所有AI Box设备的简洁列表
 * - 显示在线状态、位置等关键信息
 * - 支持搜索和筛选
 * - 点击设备切换选择
 */
const DeviceList: React.FC<DeviceListProps> = ({
  devices,
  selectedDeviceId,
  loading = false,
  onSelect,
}) => {
  const [searchText, setSearchText] = React.useState("");

  // 筛选设备列表
  const filteredDevices = React.useMemo(() => {
    if (!searchText) return devices;

    const lowerSearch = searchText.toLowerCase();
    return devices.filter(
      (device) =>
        device.deviceId.toLowerCase().includes(lowerSearch) ||
        device.deviceName.toLowerCase().includes(lowerSearch) ||
        (device.deviceAlias?.toLowerCase().includes(lowerSearch) ?? false) ||
        (device.location?.toLowerCase().includes(lowerSearch) ?? false)
    );
  }, [devices, searchText]);

  // 在线设备数量
  const onlineCount = devices.filter((d) => d.isOnline).length;

  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <Spin tip="加载设备列表...">
          <div />
        </Spin>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      {/* 搜索框 */}
      <div className={styles.searchContainer}>
        <Input
          placeholder="搜索设备..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          allowClear
        />
        <div className={styles.searchStats}>
          <Text type="secondary" className={styles.searchStatsText}>
            在线: {onlineCount} / 总数: {devices.length}
          </Text>
        </div>
      </div>

      {/* 设备列表 */}
      <div className={styles.listContainer}>
        {filteredDevices.length === 0 ? (
          <Empty
            description={searchText ? "未找到匹配设备" : "暂无AI Box设备"}
            className={styles.emptyContainer}
          />
        ) : (
          <List
            dataSource={filteredDevices}
            renderItem={(device) => {
              const isSelected = device.deviceId === selectedDeviceId;

              return (
                <List.Item
                  onClick={() => onSelect?.(device.deviceId)}
                  className={classNames(styles.listItem, {
                    [styles.selected]: isSelected,
                  })}
                >
                  <div className={styles.listItemContent}>
                    {/* 设备名称和在线状态 */}
                    <div className={styles.deviceHeader}>
                      <Text
                        strong={isSelected}
                        className={classNames(styles.deviceName, {
                          [styles.deviceNameSelected]: isSelected,
                        })}
                        ellipsis={{
                          tooltip: device.deviceAlias || device.deviceName,
                        }}
                      >
                        {device.deviceAlias || device.deviceName}
                      </Text>
                      {device.isOnline ? (
                        <CheckCircleOutlined
                          className={styles.statusIconOnline}
                        />
                      ) : (
                        <MinusCircleOutlined
                          className={styles.statusIconOffline}
                        />
                      )}
                    </div>

                    {/* 设备ID */}
                    <div className={styles.deviceIdRow}>
                      <Text type="secondary" className={styles.deviceIdText}>
                        ID: {device.deviceId}
                      </Text>
                    </div>

                    {/* 位置信息 */}
                    {device.location && (
                      <div>
                        <Space size={4}>
                          <EnvironmentOutlined
                            className={styles.locationIcon}
                          />
                          <Text
                            type="secondary"
                            className={styles.locationText}
                            ellipsis
                          >
                            {device.location}
                          </Text>
                        </Space>
                      </div>
                    )}
                  </div>
                </List.Item>
              );
            }}
          />
        )}
      </div>
    </div>
  );
};

export default DeviceList;
