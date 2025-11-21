import type { GB28181Channel, NexlynDevice } from "@/services/video";
import { getDeviceChannels, getNexlynDeviceDetail } from "@/services/video";
import {
  LinkOutlined,
  LoadingOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons";
import { Button, Empty, message, Space, Table, Tag } from "antd";
import React, { useEffect, useState } from "react";
import type { DeviceBindingFilter } from "../types";
import styles from "./DeviceChannelDetail.less";

interface DeviceChannelDetailProps {
  device?: NexlynDevice | null;
  bindingFilter: DeviceBindingFilter;
  organizationId?: string; // 当前选中的组织ID
  onPlay: (deviceId: string, channelId: string, channelName?: string) => void;
  onShowPlayUrl?: (channel: GB28181Channel) => void;
}

const getDeviceStatusColor = (status: string) => {
  switch (status) {
    case "REGIST":
      return "green";
    case "ONLINE":
      return "blue";
    case "OFFLINE":
      return "red";
    default:
      return "default";
  }
};

const getChannelStatusColor = (status: string) =>
  status === "ON" ? "success" : "default";

const DeviceChannelDetail: React.FC<DeviceChannelDetailProps> = ({
  device,
  bindingFilter,
  organizationId,
  onPlay,
  onShowPlayUrl,
}) => {
  const [deviceDetail, setDeviceDetail] = useState<any>(null);
  const [channels, setChannels] = useState<GB28181Channel[]>([]);
  const [deviceDetailLoading, setDeviceDetailLoading] = useState(false);
  const [channelLoading, setChannelLoading] = useState(false);

  // 当选中设备变化时，加载设备详情和通道信息
  useEffect(() => {
    if (!device?.deviceId) {
      setDeviceDetail(null);
      setChannels([]);
      return;
    }

    // 检查设备绑定状态，未绑定设备不允许查看详情
    if (bindingFilter === "unbound") {
      setDeviceDetail(null);
      setChannels([]);
      return;
    }

    const loadDeviceDetails = async () => {
      setDeviceDetailLoading(true);
      setChannelLoading(true);

      try {
        const deviceId = device.deviceId;

        // 并发请求设备详情和通道列表
        const [deviceDetailResponse, channelsResponse] = await Promise.all([
          getNexlynDeviceDetail(deviceId, { organizationId }),
          getDeviceChannels(deviceId, { page: 0, count: 0 }),
        ]);

        // 处理设备详情响应
        if (deviceDetailResponse.code === 0) {
          const detailData = {
            ...deviceDetailResponse.data,
            channelCount: deviceDetailResponse.data.channelCount || 0,
            channels: deviceDetailResponse.data.channels || [],
          };
          setDeviceDetail(detailData);

          // 使用设备详情API返回的通道数据，如果通道API失败的话作为备用
          if (channelsResponse.code !== 0 && detailData.channels) {
            const channelsWithDeviceId = (detailData.channels || []).map(
              (ch: GB28181Channel) => ({
                ...ch,
                deviceId: deviceId,
              })
            );
            setChannels(channelsWithDeviceId);
          }
        } else {
          message.error(deviceDetailResponse.message || "获取设备详情失败");
          setDeviceDetail(null);
        }

        // 处理通道列表响应
        if (channelsResponse.code === 0) {
          const channelsWithDeviceId = (channelsResponse.list || []).map(
            (ch: GB28181Channel) => ({
              ...ch,
              deviceId: deviceId,
            })
          );
          setChannels(channelsWithDeviceId);
        } else if (!deviceDetailResponse.data?.channels) {
          // 如果通道API失败且设备详情中也没有通道数据，显示错误
          message.error(channelsResponse.message || "获取通道详情失败");
          setChannels([]);
        }
      } catch (_error) {
        message.error("获取设备信息失败");
        setDeviceDetail(null);
        setChannels([]);
      } finally {
        setDeviceDetailLoading(false);
        setChannelLoading(false);
      }
    };

    loadDeviceDetails();
  }, [device?.deviceId, bindingFilter]);

  // 处理播放地址按钮点击
  const handleShowPlayUrl = (channel: GB28181Channel) => {
    if (onShowPlayUrl) {
      onShowPlayUrl(channel);
    } else {
      // 如果没有提供回调，显示默认的播放路径
      const playPath = `${channel.deviceId}/${channel.channelId}`;
      message.info(`播放地址: ${playPath}`);
    }
  };

  // 如果没有选中设备，显示空状态
  if (!device) {
    return (
      <div className={styles.emptyContainer}>
        <Empty description="请选择设备查看通道详情" />
      </div>
    );
  }

  // 如果是未绑定设备，显示提示信息
  if (bindingFilter === "unbound") {
    return (
      <div className={styles.emptyContainer}>
        <Empty
          description="未绑定设备无法查看通道详情"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <h3 className={styles.title}>
        设备通道详情 -{" "}
        {deviceDetail?.name || device?.name || device?.deviceId || ""}
      </h3>
      {/* 设备基本信息 */}
      <div className={styles.deviceInfo}>
        {deviceDetailLoading ? (
          <div className={styles.loadingContainer}>
            <LoadingOutlined /> 加载设备详情中...
          </div>
        ) : (
          <Space direction="vertical" className={styles.infoSpace} size="small">
            <div>
              <strong>设备ID:</strong>{" "}
              {deviceDetail?.deviceId || device?.deviceId}
            </div>
            <div>
              <strong>设备名称:</strong>{" "}
              {deviceDetail?.name || device?.name || "未命名"}
            </div>
            <div>
              <strong>制造商:</strong>{" "}
              {deviceDetail?.manufacturer || device?.manufacturer || "未知"}
            </div>
            <div>
              <strong>型号:</strong>{" "}
              {deviceDetail?.model || device?.model || "未知"}
            </div>
            <div>
              <strong>状态:</strong>
              <Tag
                color={getDeviceStatusColor(
                  deviceDetail?.status || device?.status || ""
                )}
                className={styles.statusTag}
              >
                {deviceDetail?.status || device?.status}
              </Tag>
            </div>
            {deviceDetail?.streamMode && (
              <div>
                <strong>传输模式:</strong> {deviceDetail.streamMode}
              </div>
            )}
          </Space>
        )}
      </div>

      {/* 通道列表 */}
      <div>
        <h4>通道列表</h4>
        <Table
          dataSource={channels}
          loading={channelLoading}
          rowKey="channelId"
          size="small"
          pagination={{
            pageSize: 10,
            showSizeChanger: false,
            showQuickJumper: false,
            showTotal: (total) => `共 ${total} 个通道`,
          }}
          scroll={{ y: 300, x: "max-content" }}
          columns={[
            {
              title: "通道ID",
              dataIndex: "channelId",
              key: "channelId",
              width: 100,
              ellipsis: true,
            },
            {
              title: "通道名称",
              dataIndex: "name",
              key: "name",
              ellipsis: true,
            },
            {
              title: "制造商",
              dataIndex: "manufacturer",
              key: "manufacturer",
              ellipsis: true,
            },
            {
              title: "状态",
              dataIndex: "status",
              key: "status",
              width: 60,
              render: (status: string) => (
                <Tag color={getChannelStatusColor(status)}>{status}</Tag>
              ),
            },
            {
              title: "地址",
              dataIndex: "address",
              key: "address",
              ellipsis: true,
            },
            {
              title: "操作",
              key: "action",
              width: 160,
              render: (_: any, record: GB28181Channel) => (
                <div className={styles.actionButtons}>
                  <Button
                    type="link"
                    size="small"
                    icon={<PlayCircleOutlined />}
                    onClick={() =>
                      onPlay(record.deviceId, record.channelId, record.name)
                    }
                    disabled={record.status !== "ON"}
                  >
                    预览
                  </Button>
                  <Button
                    type="link"
                    size="small"
                    icon={<LinkOutlined />}
                    onClick={() => handleShowPlayUrl(record)}
                    disabled={record.status !== "ON"}
                  >
                    播放地址
                  </Button>
                </div>
              ),
            },
          ]}
        />
      </div>
    </div>
  );
};

export default DeviceChannelDetail;
