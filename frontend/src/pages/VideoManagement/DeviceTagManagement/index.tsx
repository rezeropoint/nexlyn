import { ExclamationCircleOutlined } from "@ant-design/icons";
import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import CreateDeviceTagForm from "./components/CreateDeviceTagForm";
import DeviceTagTable from "./components/DeviceTagTable";
import EditDeviceTagForm from "./components/EditDeviceTagForm";
import { useDeviceTagManagement } from "./hooks/useDeviceTagManagement";
import { useDeviceTagPermissions } from "./hooks/useDeviceTagPermissions";
import styles from "./index.less";

/**
 * 设备标签管理主页面
 *
 * 重构后的组件结构：
 * - 主页面只负责组件编排和状态协调
 * - 具体业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const DeviceTagManagement: React.FC = () => {
  const permissions = useDeviceTagPermissions();
  const actionRef = useRef<ActionType>();

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,
    selectionState,

    // 操作方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleBatchDelete,
    handleEditClick,
  } = useDeviceTagManagement();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建设备标签成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新设备标签成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 设备标签操作处理（包含表格刷新）
  const handleDeviceTagOperation = {
    delete: async (record: any) => {
      await handleDelete(record);
      handleTableRefresh();
    },
    batchDelete: async (tags: any[]) => {
      await handleBatchDelete(tags);
      handleTableRefresh();
    },
  };

  // 权限检查
  if (!permissions.canRead) {
    return (
      <PageContainer>
        <div className={styles.permissionDenied}>
          <ExclamationCircleOutlined className={styles.icon} />
          <h2>权限不足</h2>
          <p>您需要设备标签管理权限才能访问此页面</p>
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer
      header={{
        title: "设备标签管理",
        subTitle: "Nexlyn设备标签系统管理",
      }}
    >
      {/* 主表格 */}
      <DeviceTagTable
        actionRef={actionRef}
        selectionState={selectionState}
        onSelectionChange={setSelectionState}
        onCreateTag={() => setModalOpen("create", true)}
        onEditTag={handleEditClick}
        onDeleteTag={handleDeviceTagOperation.delete}
        onBatchDelete={handleDeviceTagOperation.batchDelete}
      />

      {/* 创建设备标签表单 */}
      {permissions.canCreate && (
        <CreateDeviceTagForm
          open={modalState.create}
          onOpenChange={(open) => setModalOpen("create", open)}
          onFinish={handleCreateSuccess}
        />
      )}

      {/* 编辑设备标签表单 */}
      {permissions.canUpdate && (
        <EditDeviceTagForm
          open={modalState.edit}
          onOpenChange={(open) => setModalOpen("edit", open)}
          onFinish={handleUpdateSuccess}
          currentRow={currentRow}
        />
      )}
    </PageContainer>
  );
};

export default DeviceTagManagement;
