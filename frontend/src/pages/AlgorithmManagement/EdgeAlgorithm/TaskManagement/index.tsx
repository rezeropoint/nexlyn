import OrganizationSelector, {
  type OrganizationSelectorRef,
} from "@/pages/IoTManagement/DeviceManagement/components/OrganizationSelector";
import {
  CompressOutlined,
  ExpandAltOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { PageContainer, ProCard } from "@ant-design/pro-components";
import { Button, Card, Space, Typography, theme } from "antd";
import React, { useRef, useState } from "react";
import CapabilitiesDrawer from "./components/CapabilitiesDrawer";
import DeviceInfoCard from "./components/DeviceInfoCard";
import DeviceList from "./components/DeviceList";
import TaskTable from "./components/TaskTable";
import { useAIBoxControl } from "./hooks/useAIBoxControl";
import styles from "./index.less";

const { Paragraph } = Typography;

/**
 * 任务管理页面
 *
 * 功能：
 * - AI Box设备选择
 * - 设备信息展示
 * - 算法任务列表查询
 * - 算法能力查询
 *
 * 布局（三栏布局）：
 * - 左侧：组织树选择器（20%）
 * - 中间：设备列表（18%）
 * - 右侧：设备信息卡片 + 任务列表表格（62%）
 */
const TaskManagement: React.FC = () => {
  const { token } = theme.useToken();
  const {
    // 状态
    selectedOrgId,
    deviceOptions,
    selectedDeviceId,
    deviceLoading,
    tasks,
    tasksLoading,
    capabilities,
    capabilitiesLoading,
    drawerOpen,

    // 方法
    handleOrgSelect,
    handleDeviceChange,
    handleOpenCapabilities,
    handleCloseCapabilities,
    refreshTasks,
    handleControlTask,
  } = useAIBoxControl();

  // 组织树的 ref
  const orgSelectorRef = useRef<OrganizationSelectorRef>(null);
  const [orgTreeLoading, setOrgTreeLoading] = useState(false);

  // 获取当前选中的设备详细信息
  const currentDevice =
    deviceOptions.find((d) => d.deviceId === selectedDeviceId) || null;

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
          // 300ms 后重置 loading 状态
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
        title: "任务管理",
        subTitle: "AI Box算法任务查询与能力管理",
      }}
    >
      {/* 三栏布局 */}
      <ProCard split="vertical" gutter={[16, 16]}>
        {/* 左侧：组织树（20%） */}
        <ProCard colSpan="20%" title="组织架构" extra={orgToolbar}>
          <OrganizationSelector
            ref={orgSelectorRef}
            selectedOrgId={selectedOrgId}
            onSelect={handleOrgSelect}
            hideToolbar={true}
          />
        </ProCard>

        {/* 中间：设备列表（18%） */}
        <ProCard
          colSpan="18%"
          title={
            <Space>
              <span>设备列表</span>
              {selectedOrgId && deviceOptions.length > 0 && (
                <span
                  style={{
                    fontSize: 12,
                    fontWeight: "normal",
                    color: token.colorTextTertiary,
                  }}
                >
                  ({deviceOptions.filter((d) => d.isOnline).length} 在线)
                </span>
              )}
            </Space>
          }
          className={styles.deviceListCard}
        >
          <DeviceList
            devices={deviceOptions}
            selectedDeviceId={selectedDeviceId}
            loading={deviceLoading}
            onSelect={handleDeviceChange}
          />
        </ProCard>

        {/* 右侧：设备信息和任务列表（62%） */}
        <ProCard colSpan="62%">
          <Space direction="vertical" size="middle" style={{ width: "100%" }}>
            {/* 设备信息卡片 */}
            <Card className={styles.deviceInfoCard}>
              <DeviceInfoCard
                device={currentDevice}
                onViewCapabilities={handleOpenCapabilities}
              />
            </Card>

            {/* 使用说明（当没有选择组织或设备时显示） */}
            {!selectedOrgId && (
              <Card>
                <Paragraph>
                  <strong>使用说明：</strong>
                </Paragraph>
                <ol>
                  <li>
                    在左侧选择组织架构，中间将显示该组织下的AI Box设备列表
                  </li>
                  <li>在中间设备列表中点击要管理的AI Box设备</li>
                  <li>右侧将显示该设备的详细信息和算法任务列表</li>
                  <li>点击"查看算法能力"按钮可查看设备支持的所有算法详情</li>
                </ol>
              </Card>
            )}

            {/* 任务列表表格 */}
            <TaskTable
              tasks={tasks}
              loading={tasksLoading}
              deviceId={selectedDeviceId}
              isOnline={currentDevice?.isOnline || false}
              capabilities={capabilities}
              onRefresh={refreshTasks}
              onControlTask={handleControlTask}
            />
          </Space>
        </ProCard>
      </ProCard>

      {/* 算法能力抽屉 */}
      <CapabilitiesDrawer
        open={drawerOpen}
        deviceId={selectedDeviceId}
        capabilities={capabilities}
        loading={capabilitiesLoading}
        onClose={handleCloseCapabilities}
      />
    </PageContainer>
  );
};

export default TaskManagement;
