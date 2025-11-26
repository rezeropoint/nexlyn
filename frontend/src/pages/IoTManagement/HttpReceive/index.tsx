import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import CreateHttpReceiveForm from "./components/CreateHttpReceiveForm";
import EditHttpReceiveForm from "./components/EditHttpReceiveForm";
import HttpReceiveTable from "./components/HttpReceiveTable";
import { useHttpReceive } from "./hooks/useHttpReceive";

/**
 * HTTP 接收配置管理主页面
 *
 * 组件结构：
 * - 主页面负责组件编排和状态协调
 * - 业务逻辑由自定义 Hook 管理
 * - UI 组件拆分为可复用的子组件
 */
const HttpReceiveManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();

  // 使用自定义 Hook 管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,

    // 设置方法
    setModalOpen,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleEditClick,
    handleToggleEnabled,
  } = useHttpReceive();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建配置成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新配置成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 删除配置后刷新表格
  const handleDeleteConfig = async (record: any) => {
    await handleDelete(record);
    handleTableRefresh();
  };

  return (
    <PageContainer subTitle="配置 HTTP 数据接收端点，适用于无法通过 MQTT 提供数据的设备">
      {/* 主表格 */}
      <HttpReceiveTable
        actionRef={actionRef}
        onCreateConfig={() => setModalOpen("create", true)}
        onEditConfig={handleEditClick}
        onDeleteConfig={handleDeleteConfig}
        onToggleEnabled={handleToggleEnabled}
      />

      {/* 创建配置表单 */}
      <CreateHttpReceiveForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen("create", open)}
        onFinish={handleCreateSuccess}
      />

      {/* 编辑配置表单 */}
      <EditHttpReceiveForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        initialData={currentRow}
      />
    </PageContainer>
  );
};

export default HttpReceiveManagement;
