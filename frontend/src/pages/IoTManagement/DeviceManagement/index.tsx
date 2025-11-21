import {
  CompressOutlined,
  ExpandAltOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import {
  type ActionType,
  PageContainer,
  ProCard,
} from "@ant-design/pro-components";
import { Button, Space, Splitter, Tag } from "antd";
import React, { useEffect, useRef, useState } from "react";
import useResponsiveLayout from "@/hooks/useResponsiveLayout";
import { getUnifiedDeviceStatistics } from "@/services/device";
import BindDeviceForm from "./components/BindDeviceForm";
import DeviceDetailDrawer from "./components/DeviceDetailDrawer";
import DeviceTable from "./components/DeviceTable";
import EditDeviceForm from "./components/EditDeviceForm";
import OrganizationSelector, {
  type OrganizationSelectorRef,
} from "./components/OrganizationSelector";
import { useDeviceBinding } from "./hooks/useDeviceBinding";
import styles from "./index.less";

/**
 * 设备管理主页面
 *
 * 组件结构：
 * - 左右分栏布局：左侧组织树选择器，右侧设备列表
 * - 主页面负责组件编排和状态协调
 * - 业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const DeviceManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const orgSelectorRef = useRef<OrganizationSelectorRef>(null);
  const layout = useResponsiveLayout({ breakpoint: "xl" });
  const [selectedOrgId, setSelectedOrgId] = useState<string>("");
  const [orgTreeLoading, setOrgTreeLoading] = useState(false);

  // IoT设备统计数据
  const [deviceSummary, setDeviceSummary] = useState({
    total: 0,
    online: 0,
    offline: 0,
  });

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    currentDevice,
    modalState,
    selectionState,

    // 设置方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleBind,
    handleUpdate,
    handleUnbind,
    handleBatchUnbind,
    handleViewDetail,
    handleEditClick,
  } = useDeviceBinding();

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

  // 更新设备成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 设备操作处理（包含表格刷新）
  const handleDeviceOperation = {
    unbind: async (record: any) => {
      await handleUnbind(record);
      handleTableRefresh();
    },
    batchUnbind: async (devices: any[]) => {
      await handleBatchUnbind(devices);
      handleTableRefresh();
    },
  };

  // 处理组织选择变化
  const handleOrgSelect = (orgId: string) => {
    setSelectedOrgId(orgId);
    // 刷新设备列表
    actionRef.current?.reload();
  };

  // 加载设备统计数据
  useEffect(() => {
    const loadStats = async () => {
      // 只有在选择了组织时才加载统计
      if (!selectedOrgId) {
        return;
      }

      try {
        const data = await getUnifiedDeviceStatistics({
          organizationId: selectedOrgId,
          deviceType: "iot_sensor", // 只统计IoT传感器设备
        });

        setDeviceSummary({
          total: data.deviceTotal || 0,
          online: data.deviceOnline || 0,
          offline: data.deviceOffline || 0,
        });
      } catch (error) {
        // 统计接口失败时优雅降级，不影响主要功能
        console.error("加载IoT设备统计失败:", error);
        // 重置为0，避免显示旧数据
        setDeviceSummary({ total: 0, online: 0, offline: 0 });
      }
    };

    loadStats();
  }, [selectedOrgId]);

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

  return (
    <PageContainer
      header={{
        title: "设备管理",
        subTitle: "管理物联网设备绑定、监控和配置",
        extra: [
          <Space key="stats">
            <Tag color="blue">设备: {deviceSummary.total}</Tag>
            <Tag color="success">在线: {deviceSummary.online}</Tag>
            <Tag color="default">离线: {deviceSummary.offline}</Tag>
          </Space>,
        ],
      }}
    >
      {/* 左右分栏布局 */}
      <Splitter layout={layout} className={styles.splitterContainer}>
        {/* 左侧组织树 */}
        <Splitter.Panel
          defaultSize="25%"
          min={180}
          max="45%"
          collapsible={{ start: true }}
          className={styles.leftPanel}
        >
          <ProCard title="组织架构" extra={orgToolbar}>
            <OrganizationSelector
              ref={orgSelectorRef}
              selectedOrgId={selectedOrgId}
              onSelect={handleOrgSelect}
              hideToolbar={true}
            />
          </ProCard>
        </Splitter.Panel>

        {/* 右侧设备列表 */}
        <Splitter.Panel min={600} max="80%" className={styles.rightPanel}>
          <ProCard>
            <DeviceTable
              actionRef={actionRef}
              selectedOrgId={selectedOrgId}
              selectionState={selectionState}
              onSelectionChange={(state) =>
                setSelectionState((prev) => ({ ...prev, ...state }))
              }
              onBindDevice={() => setModalOpen("bind", true)}
              onViewDevice={handleViewDetail}
              onEditDevice={handleEditClick}
              onUnbindDevice={handleDeviceOperation.unbind}
              onBatchUnbind={handleDeviceOperation.batchUnbind}
            />
          </ProCard>
        </Splitter.Panel>
      </Splitter>

      {/* 绑定设备表单 */}
      <BindDeviceForm
        open={modalState.bind}
        onOpenChange={(open) => setModalOpen("bind", open)}
        onFinish={handleBindSuccess}
      />

      {/* 编辑设备表单 */}
      <EditDeviceForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />

      {/* 设备详情抽屉 */}
      <DeviceDetailDrawer
        open={modalState.detail}
        onClose={() => setModalOpen("detail", false)}
        device={currentDevice}
      />
    </PageContainer>
  );
};

export default DeviceManagement;
