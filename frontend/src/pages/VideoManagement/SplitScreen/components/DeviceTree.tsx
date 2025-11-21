import type { NexlynDevice } from "@/services/video";
import { getDeviceChannels, getNexlynDeviceList } from "@/services/video";
import {
  DesktopOutlined,
  DragOutlined,
  ReloadOutlined,
  SearchOutlined,
  VideoCameraOutlined,
} from "@ant-design/icons";
import { Button, Input, message, Pagination, Space, Spin, Tree } from "antd";
import classNames from "classnames";
import React, { useCallback, useEffect, useState } from "react";
import OrganizationSelector from "../../components/OrganizationSelector";
import type { ChannelPlayInfo, DeviceTreeNode } from "../types";
import styles from "./DeviceTree.less";

const { Search } = Input;

interface DeviceTreeProps {
  onChannelPlay: (gridIndex: number, channelInfo: ChannelPlayInfo) => void;
  playingChannels?: Map<number, ChannelPlayInfo>;
  maxGrids?: number;
  organizationId?: string;
  onOrganizationChange?: (organizationId: string | undefined) => void;
}

const DeviceTree: React.FC<DeviceTreeProps> = ({
  onChannelPlay,
  playingChannels = new Map(),
  maxGrids = 16,
  organizationId,
  onOrganizationChange,
}) => {
  const [devices, setDevices] = useState<NexlynDevice[]>([]);
  const [treeData, setTreeData] = useState<DeviceTreeNode[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [searchValue, setSearchValue] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadingNodes, setLoadingNodes] = useState<Set<string>>(new Set());
  const [currentPage, setCurrentPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;

  // 加载设备列表
  const loadDevices = useCallback(
    async (page = currentPage) => {
      // 如果组织ID未加载完成，跳过请求
      if (organizationId === undefined) {
        return;
      }

      setLoading(true);
      try {
        const response = await getNexlynDeviceList({
          page,
          pageSize,
          bindStatus: "bound", // 只获取已绑定的设备
          keyword: searchValue || undefined,
          organizationId: organizationId, // 组织筛选
        });

        if (response.code === 0) {
          const deviceList = response.data?.devices || [];
          setDevices(deviceList);
          setTotal(response.data?.total || 0);
          generateTreeData(deviceList);
        } else {
          message.error(response.message || "获取设备列表失败");
        }
      } catch (error) {
        console.error("加载设备列表失败:", error);
        message.error("加载设备列表失败");
      } finally {
        setLoading(false);
      }
    },
    [currentPage, pageSize, searchValue, organizationId]
  );

  // 生成树形数据
  const generateTreeData = useCallback((deviceList: NexlynDevice[]) => {
    // 后端已经处理了搜索过滤，这里直接使用
    const filteredDevices = deviceList;

    const treeNodes: DeviceTreeNode[] = filteredDevices.map((device) => {
      const deviceName = device.deviceAlias || device.name || device.deviceId;
      const isOnline = device.online || device.status === "ON";

      return {
        key: `device-${device.deviceId}`,
        title: (
          <Space>
            <DesktopOutlined
              className={classNames(styles.deviceIcon, {
                [styles.online]: isOnline,
                [styles.offline]: !isOnline,
              })}
            />
            <span>{deviceName}</span>
          </Space>
        ),
        deviceId: device.deviceId,
        deviceName,
        status: device.status,
        isLeaf: false,
        children: [],
      };
    });

    setTreeData(treeNodes);
  }, []);

  // 加载设备通道
  const loadDeviceChannels = useCallback(
    async (deviceId: string) => {
      const loadingKey = `device-${deviceId}`;
      setLoadingNodes((prev) => new Set(prev.add(loadingKey)));

      try {
        const response = await getDeviceChannels(deviceId, {
          page: 0,
          count: 0, // 获取所有通道
        });

        if (response.code === 0) {
          const channels = response.list || [];
          const device = devices.find((d) => d.deviceId === deviceId);
          const deviceName = device?.deviceAlias || device?.name || deviceId;

          // 生成通道节点
          const channelNodes: DeviceTreeNode[] = channels.map((channel) => {
            const isOnline = channel.status === "ON";
            return {
              key: `channel-${deviceId}-${channel.channelId}`,
              title: (
                <div
                  draggable={isOnline}
                  onDragStart={(e) => {
                    if (isOnline) {
                      const channelData = {
                        deviceId,
                        channelId: channel.channelId,
                        channelName: channel.name,
                        deviceName,
                      };
                      e.dataTransfer.setData(
                        "text/plain",
                        JSON.stringify(channelData)
                      );
                      e.dataTransfer.effectAllowed = "copy";
                    }
                  }}
                  className={classNames(styles.channelNode, {
                    [styles.offline]: !isOnline,
                  })}
                >
                  <Space>
                    {isOnline && <DragOutlined className={styles.dragIcon} />}
                    <VideoCameraOutlined
                      className={classNames(styles.channelIcon, {
                        [styles.online]: isOnline,
                        [styles.offline]: !isOnline,
                      })}
                    />
                    <span>{channel.name || channel.channelId}</span>
                  </Space>
                </div>
              ),
              deviceId,
              channelId: channel.channelId,
              channelName: channel.name,
              deviceName,
              status: channel.status,
              isLeaf: true,
              isChannel: true,
            };
          });

          // 更新树形数据
          setTreeData((prevTreeData) =>
            prevTreeData.map((node) => {
              if (node.key === loadingKey) {
                return {
                  ...node,
                  children: channelNodes,
                };
              }
              return node;
            })
          );

          // 自动展开该设备节点
          setExpandedKeys((prev) => [...prev, loadingKey]);
        } else {
          message.error(`获取设备 ${deviceId} 的通道失败: ${response.message}`);
        }
      } catch (error) {
        console.error("加载设备通道失败:", error);
        message.error("加载设备通道失败");
      } finally {
        setLoadingNodes((prev) => {
          const newSet = new Set(prev);
          newSet.delete(loadingKey);
          return newSet;
        });
      }
    },
    [devices]
  );

  // 处理通道双击播放
  const handleChannelDoubleClick = useCallback(
    (
      deviceId: string,
      channelId: string,
      channelName?: string,
      deviceName?: string
    ) => {
      // 找到一个空闲的网格位置
      let emptyGridIndex = -1;
      for (let i = 0; i < maxGrids; i++) {
        if (!playingChannels.has(i)) {
          emptyGridIndex = i;
          break;
        }
      }

      if (emptyGridIndex === -1) {
        message.warning("所有网格都已被占用，请先停止一个视频播放");
        return;
      }

      onChannelPlay(emptyGridIndex, {
        deviceId,
        channelId,
        channelName,
        deviceName,
      });
      message.success(
        `开始播放: ${channelName || channelId} (网格 ${emptyGridIndex + 1})`
      );
    },
    [onChannelPlay, playingChannels, maxGrids]
  );

  // 处理树节点展开
  const handleTreeExpand = useCallback(
    (expandedKeys: React.Key[], info: any) => {
      setExpandedKeys(expandedKeys);

      // 如果是展开设备节点且还没有加载通道，则加载通道
      if (info.expanded && info.node.deviceId && !info.node.isChannel) {
        const hasChildren = info.node.children && info.node.children.length > 0;
        if (!hasChildren) {
          loadDeviceChannels(info.node.deviceId);
        }
      }
    },
    [loadDeviceChannels]
  );

  // 处理搜索
  const handleSearch = useCallback((value: string) => {
    setSearchValue(value);
    setCurrentPage(1); // 搜索时重置到第一页
  }, []);

  // 处理分页变化
  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

  // 初始化加载
  useEffect(() => {
    loadDevices();
  }, [loadDevices]);

  // 当页码、搜索条件或组织筛选变化时重新加载
  useEffect(() => {
    loadDevices(currentPage);
  }, [currentPage, searchValue, organizationId]);

  return (
    <div className={styles.deviceTree}>
      {/* 组织筛选、搜索和刷新 */}
      <Space direction="vertical" className={styles.controls}>
        <OrganizationSelector
          value={organizationId}
          onChange={onOrganizationChange}
          placeholder="选择组织架构"
          style={{ width: "100%" }}
        />
        <Search
          placeholder="搜索设备"
          allowClear
          prefix={<SearchOutlined />}
          onChange={(e) => handleSearch(e.target.value)}
          style={{ width: "100%" }}
        />
        <Button
          icon={<ReloadOutlined />}
          onClick={() => loadDevices()}
          loading={loading}
          className={styles.refreshBtn}
        >
          刷新设备列表
        </Button>
      </Space>

      {/* 设备树 */}
      <div className={styles.treeContainer}>
        <Spin spinning={loading} tip="加载设备列表...">
          <Tree
            treeData={treeData}
            expandedKeys={expandedKeys}
            onExpand={handleTreeExpand}
            showIcon
            blockNode
            style={{ background: "transparent" }}
            titleRender={(nodeData) => {
              const isLoading = loadingNodes.has(nodeData.key as string);
              return (
                <div className={styles.treeNodeWrapper}>
                  {isLoading && (
                    <Spin size="small" style={{ marginRight: 8 }} />
                  )}
                  {nodeData.title}
                </div>
              );
            }}
            onDoubleClick={(_e, node) => {
              // 双击通道节点播放视频
              if (
                node.isChannel &&
                node.deviceId &&
                node.channelId &&
                node.status === "ON"
              ) {
                handleChannelDoubleClick(
                  node.deviceId,
                  node.channelId,
                  node.channelName,
                  node.deviceName
                );
              }
            }}
          />
        </Spin>
      </div>

      {/* 分页器 */}
      {total > pageSize && (
        <div className={styles.pagination}>
          <Pagination
            current={currentPage}
            total={total}
            pageSize={pageSize}
            onChange={handlePageChange}
            showSizeChanger={false}
            showQuickJumper={false}
            showTotal={(total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条设备`
            }
            size="small"
          />
        </div>
      )}
    </div>
  );
};

export default DeviceTree;
