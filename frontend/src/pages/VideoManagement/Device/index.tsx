import VideoPlayer from "@/components/VideoPlayer";
import OrganizationSelector, {
  type OrganizationSelectorRef,
} from "@/pages/IoTManagement/DeviceManagement/components/OrganizationSelector";
import type { GB28181Channel, NexlynDevice } from "@/services/video";
import {
  CompressOutlined,
  ExpandAltOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { ActionType } from "@ant-design/pro-components";
import { PageContainer, ProCard } from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { Badge, Button, Space, Splitter, Tag, Tooltip } from "antd";
import React, { useEffect, useRef, useState } from "react";
import useResponsiveLayout from "@/hooks/useResponsiveLayout";
import DeviceTable from "../components/DeviceTable";
import PlayUrlModal from "../Media/components/PlayUrlModal";
import DeviceBindForm from "./components/DeviceBindForm";
import DeviceChannelDetail from "./components/DeviceChannelDetail";
import EditDeviceAliasForm from "./components/EditDeviceAliasForm";
import EditDeviceForm from "./components/EditDeviceForm";
import EditDeviceTagsForm from "./components/EditDeviceTagsForm";
import { useDeviceManagement } from "./hooks/useDeviceManagement";
import { useDevicePermissions } from "./hooks/useDevicePermissions";
import styles from "./index.less";
import type { DeviceBindingFilter, DeviceSummaryStats } from "./types";

const DeviceManagement: React.FC = () => {
  const permissions = useDevicePermissions();
  const actionRef = useRef<ActionType>();
  const orgSelectorRef = useRef<OrganizationSelectorRef>(null);
  const layout = useResponsiveLayout({ breakpoint: "xxl" });
  const { initialState } = useModel("@@initialState");

  // 组织架构筛选（必须在 useDeviceManagement 之前声明）
  const [organizationId, setOrganizationId] = useState<string>();
  // 区分"首次初始化加载"和"用户清空选择"的状态
  const [orgInitDone, setOrgInitDone] = useState(false);
  const [orgTreeLoading, setOrgTreeLoading] = useState(false);

  // 使用设备管理Hook
  const {
    // 状态
    currentDevice,
    modalState,
    availableTags,
    loadingUnbound,

    // 操作方法
    setModalOpen,

    // 业务操作
    handleBind,
    handleUnbind,
    handleUpdateAlias,
    handleUpdateTags,
    handleEditDevice,
    handleEditClick,
    handleEditAliasClick,
    handleEditTagsClick,
  } = useDeviceManagement(organizationId);

  // 选中设备状态
  const [selectedDevice, setSelectedDevice] = useState<NexlynDevice | null>(
    null
  );

  // 其他状态
  const [autoRefresh, setAutoRefresh] = useState(false); // 默认关闭自动刷新
  const [deviceSummary, setDeviceSummary] = useState<DeviceSummaryStats>({
    deviceTotal: 0,
    deviceOnline: 0,
    deviceOffline: 0,
  });

  // 设备绑定状态筛选
  const [bindingFilter, setBindingFilter] =
    useState<DeviceBindingFilter>("bound");

  // 视频播放器状态
  const [videoPlayerVisible, setVideoPlayerVisible] = useState(false);
  const [playingChannel, setPlayingChannel] = useState<{
    deviceId: string;
    channelId: string;
    channelName?: string;
  } | null>(null);

  // 播放地址模态框状态
  const [playUrlModalVisible, setPlayUrlModalVisible] = useState(false);
  const [selectedPlayPath, setSelectedPlayPath] = useState<string>("");

  // 获取用户默认组织（使用 organizationIds 的第一个值）
  useEffect(() => {
    try {
      const orgIds = initialState?.currentUser?.organizationIds;
      if (orgIds && orgIds.length > 0) {
        setOrganizationId((prev) => prev ?? orgIds[0]);
      }
    } catch (error) {
      console.error("获取默认组织失败:", error);
    } finally {
      // 无论是否找到默认组织，首次初始化完成
      setOrgInitDone(true);
    }
  }, [initialState?.currentUser]);

  // 组织筛选变化时刷新设备列表
  useEffect(() => {
    if (organizationId !== undefined) {
      actionRef.current?.reload();
    }
  }, [organizationId]);

  // 定时轮询刷新
  useEffect(() => {
    if (!autoRefresh) return;
    const timer = setInterval(() => {
      actionRef.current?.reload();
    }, 5000);
    return () => clearInterval(timer);
  }, [autoRefresh]);

  // 播放通道视频
  const handlePlayChannel = (
    deviceId: string,
    channelId: string,
    channelName?: string
  ) => {
    setPlayingChannel({ deviceId, channelId, channelName });
    setVideoPlayerVisible(true);
  };

  // 关闭视频播放器
  const handleCloseVideoPlayer = () => {
    setVideoPlayerVisible(false);
    setPlayingChannel(null);
  };

  // 处理设备行点击
  const handleDeviceRowClick = (device: NexlynDevice) => {
    setSelectedDevice(device);
  };

  // 处理播放地址显示
  const handleShowPlayUrl = (channel: GB28181Channel) => {
    const playPath = `${channel.deviceId}/${channel.channelId}`;
    setSelectedPlayPath(playPath);
    setPlayUrlModalVisible(true);
  };

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 绑定设备成功后的回调
  const handleBindSuccess = async (values: any): Promise<boolean> => {
    const success = await handleBind(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新设备别名成功后的回调
  const handleUpdateAliasSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdateAlias(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新设备标签成功后的回调
  const handleUpdateTagsSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdateTags(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 统一编辑设备成功后的回调
  const handleEditDeviceSuccess = async (values: any): Promise<boolean> => {
    const success = await handleEditDevice(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 设备操作处理（包含表格刷新）
  const handleDeviceOperation = {
    unbind: async (device: NexlynDevice) => {
      await handleUnbind(device);
      handleTableRefresh();
    },
  };

  // 组织树工具栏
  const orgToolbar = (
    <Space size={4}>
      <Button
        size="small"
        icon={<ExpandAltOutlined />}
        onClick={() => orgSelectorRef.current?.expandAll()}
        title="全部展开"
      />
      <Button
        size="small"
        icon={<CompressOutlined />}
        onClick={() => orgSelectorRef.current?.collapseAll()}
        title="全部收起"
      />
      <Button
        size="small"
        icon={<ReloadOutlined />}
        onClick={() => {
          setOrgTreeLoading(true);
          orgSelectorRef.current?.refresh();
          setTimeout(() => setOrgTreeLoading(false), 300);
        }}
        loading={orgTreeLoading}
        title="刷新"
      />
    </Space>
  );

  // 权限检查
  if (!permissions.canRead) {
    return (
      <PageContainer>
        <div className="nl-device-center">
          <h2>权限不足</h2>
          <p>您需要设备管理权限才能访问此页面</p>
        </div>
      </PageContainer>
    );
  }

  // 设备筛选选项
  const getBindingFilterOptions = () => {
    const options = [
      { label: "已绑定设备", value: "bound" as DeviceBindingFilter },
    ];

    // 只有超级管理员可以查看未绑定设备
    if (permissions.canViewUnbound) {
      options.push({
        label: "未绑定设备",
        value: "unbound" as DeviceBindingFilter,
      });
    }

    return options;
  };

  return (
    <PageContainer
      header={{
        title: "设备管理",
        subTitle: "Nexlyn设备管理与标签系统",
        extra: [
          <Space key="extra">
            <Tooltip title={autoRefresh ? "自动刷新中" : "已暂停自动刷新"}>
              <Badge
                status={autoRefresh ? "processing" : "default"}
                text={autoRefresh ? "自动刷新" : "手动刷新"}
              />
            </Tooltip>
            <Tag color="blue">设备: {deviceSummary.deviceTotal}</Tag>
            <Tag color="success">在线: {deviceSummary.deviceOnline}</Tag>
            <Tag color="default">离线: {deviceSummary.deviceOffline}</Tag>
            <Button onClick={() => setAutoRefresh((v) => !v)}>
              {autoRefresh ? "暂停自动刷新" : "恢复自动刷新"}
            </Button>
          </Space>,
        ],
      }}
    >
      {/* 首次初始化未完成前显示加载；初始化完成后即使 organizationId 为空也正常展示 */}
      {!orgInitDone ? (
        <ProCard>
          <div className="nl-device-center">
            <p>正在加载组织架构...</p>
          </div>
        </ProCard>
      ) : (
        <>
          {/* 三栏布局 */}
          <Splitter layout={layout} className={styles.splitterContainer}>
            {/* 左侧：组织架构树 */}
            <Splitter.Panel
              defaultSize="20%"
              min={180}
              max="40%"
              collapsible={{ start: true }}
              className={styles.leftPanel}
            >
              <ProCard title="组织架构" extra={orgToolbar}>
                <OrganizationSelector
                  ref={orgSelectorRef}
                  selectedOrgId={organizationId}
                  onSelect={setOrganizationId}
                  hideToolbar={true}
                />
              </ProCard>
            </Splitter.Panel>

            {/* 中间和右侧：设备列表+详情面板 */}
            <Splitter.Panel min={800} max="85%">
              <Splitter layout={layout} className={styles.innerSplitter}>
                {/* 中间：设备列表 */}
                <Splitter.Panel
                  defaultSize="50%"
                  min={480}
                  max="70%"
                  className={styles.middlePanel}
                >
                  <ProCard>
                    <DeviceTable
                      actionRef={actionRef as any}
                      onRowClick={handleDeviceRowClick}
                      selectedDeviceId={
                        selectedDevice?.deviceId || selectedDevice?.device_id
                      }
                      onSummaryChange={setDeviceSummary}
                      onEdit={handleEditClick}
                      onEditAlias={handleEditAliasClick}
                      onEditTags={handleEditTagsClick}
                      onUnbind={handleDeviceOperation.unbind}
                      bindingFilter={bindingFilter}
                      actionConfig={{
                        showBind: false, // 不在操作列显示绑定按钮
                        showUnbind:
                          bindingFilter !== "unbound" && permissions.canUnbind,
                        showEdit: bindingFilter !== "unbound" && permissions.canEdit, // 使用统一编辑
                        showEditAlias: false, // 禁用单独的别名编辑按钮
                        showEditTags: false, // 禁用单独的标签编辑按钮
                        showSync: true,
                        showDelete: true,
                        showChannels: false, // 不需要详情按钮了
                      }}
                      // 工具栏绑定按钮
                      onBindDevice={() => setModalOpen("bind", true)}
                      bindLoading={loadingUnbound}
                      showBindButton={permissions.canBind}
                      // 筛选器
                      onBindingFilterChange={setBindingFilter}
                      bindingFilterOptions={getBindingFilterOptions()}
                      // 组织架构筛选（内部不再显示，由左侧栏控制）
                      organizationId={organizationId}
                      hideOrgSelector={true}
                    />
                  </ProCard>
                </Splitter.Panel>

                {/* 右侧：设备通道详情 */}
                <Splitter.Panel
                  min={520}
                  max="70%"
                  className={styles.rightPanel}
                >
                  <ProCard>
                    <DeviceChannelDetail
                      device={selectedDevice}
                      bindingFilter={bindingFilter}
                      organizationId={organizationId}
                      onPlay={handlePlayChannel}
                      onShowPlayUrl={handleShowPlayUrl}
                    />
                  </ProCard>
                </Splitter.Panel>
              </Splitter>
            </Splitter.Panel>
          </Splitter>

          {/* 设备绑定表单 */}
          {permissions.canBind && (
            <DeviceBindForm
              open={modalState.bind}
              onOpenChange={(open) => setModalOpen("bind", open)}
              onFinish={handleBindSuccess}
              availableTags={availableTags}
            />
          )}

          {/* 编辑设备别名表单 */}
          {permissions.canEdit && (
            <EditDeviceAliasForm
              open={modalState.editAlias}
              onOpenChange={(open) => setModalOpen("editAlias", open)}
              onFinish={handleUpdateAliasSuccess}
              currentDevice={currentDevice}
            />
          )}

          {/* 编辑设备标签表单 */}
          {permissions.canEdit && (
            <EditDeviceTagsForm
              open={modalState.editTags}
              onOpenChange={(open) => setModalOpen("editTags", open)}
              onFinish={handleUpdateTagsSuccess}
              currentDevice={currentDevice}
              availableTags={availableTags}
            />
          )}

          {/* 统一设备编辑表单 */}
          {permissions.canEdit && (
            <EditDeviceForm
              open={modalState.edit}
              onOpenChange={(open) => setModalOpen("edit", open)}
              onFinish={handleEditDeviceSuccess}
              currentDevice={currentDevice}
              availableTags={availableTags}
            />
          )}

          {/* 播放地址模态框 */}
          <PlayUrlModal
            open={playUrlModalVisible}
            path={selectedPlayPath}
            onClose={() => setPlayUrlModalVisible(false)}
          />

          {/* 视频播放器 */}
          {playingChannel && (
            <VideoPlayer
              deviceId={playingChannel.deviceId}
              channelId={playingChannel.channelId}
              channelName={playingChannel.channelName}
              visible={videoPlayerVisible}
              onClose={handleCloseVideoPlayer}
              mode="overlay"
            />
          )}
        </>
      )}
    </PageContainer>
  );
};

export default DeviceManagement;
