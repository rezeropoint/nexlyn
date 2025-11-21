// 通道播放信息
export interface ChannelPlayInfo {
  deviceId: string;
  channelId: string;
  channelName?: string;
  deviceName?: string;
}

// 设备树节点数据
export interface DeviceTreeNode {
  key: string;
  title: React.ReactNode;
  children?: DeviceTreeNode[];
  isLeaf?: boolean;
  deviceId?: string;
  channelId?: string;
  channelName?: string;
  deviceName?: string;
  status?: string;
  isChannel?: boolean;
}
