import type { GB28181Channel, GB28181Device } from "@/services/video";
import { PlayCircleOutlined } from "@ant-design/icons";
import { Button, Modal, Space, Table, Tag } from "antd";
import React from "react";
import styles from "./ChannelDetailModal.less";

interface ChannelDetailModalProps {
  open: boolean;
  onClose: () => void;
  device?: GB28181Device | null;
  deviceDetail?: GB28181Device | null;
  deviceDetailLoading?: boolean;
  channels: GB28181Channel[];
  channelsLoading?: boolean;
  onPlay: (deviceId: string, channelId: string, channelName?: string) => void;
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

const ChannelDetailModal: React.FC<ChannelDetailModalProps> = ({
  open,
  onClose,
  device,
  deviceDetail,
  deviceDetailLoading,
  channels,
  channelsLoading,
  onPlay,
}) => {
  return (
    <Modal
      title={`设备通道详情 - ${
        deviceDetail?.name || device?.name || device?.deviceId || ""
      }`}
      open={open}
      onCancel={onClose}
      footer={null}
      width={1200}
      destroyOnHidden
    >
      {(deviceDetail || device) && (
        <div>
          <div className={styles.deviceInfo}>
            {deviceDetailLoading ? (
              <div className={styles.loadingContainer}>加载设备详情中...</div>
            ) : (
              <Space
                direction="vertical"
                className={styles.infoSpace}
                size="small"
              >
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

          <div>
            <h4>通道列表</h4>
            <Table
              dataSource={channels}
              loading={channelsLoading}
              rowKey="channelId"
              size="small"
              pagination={{
                pageSize: 5,
                showSizeChanger: false,
                showQuickJumper: false,
              }}
              columns={[
                {
                  title: "通道ID",
                  dataIndex: "channelId",
                  key: "channelId",
                  width: 180,
                },
                {
                  title: "通道名称",
                  dataIndex: "name",
                  key: "name",
                  width: 150,
                },
                {
                  title: "制造商",
                  dataIndex: "manufacturer",
                  key: "manufacturer",
                  width: 120,
                },
                {
                  title: "状态",
                  dataIndex: "status",
                  key: "status",
                  width: 80,
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
                  width: 100,
                  render: (_: any, record: GB28181Channel) => (
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
                  ),
                },
              ]}
            />
          </div>
        </div>
      )}
    </Modal>
  );
};

export default ChannelDetailModal;
