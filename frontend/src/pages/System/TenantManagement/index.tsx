import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";
import CreateTenantForm from "./components/CreateTenantForm";
import EditTenantForm from "./components/EditTenantForm";
import StatusManageModal from "./components/StatusManageModal";
import TenantTable from "./components/TenantTable";
import { useTenantManagement } from "./hooks/useTenantManagement";
import { useTenantOptions } from "./hooks/useTenantOptions";

/**
 * 租户管理主页面
 *
 * 重构后的组件结构：
 * - 主页面只负责组件编排和状态协调
 * - 具体业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const TenantManagement: React.FC = () => {
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
    handleUpdateStatus,
    handleBatchDelete,
    handleBatchToggleStatus,
    handleToggleStatus,
    handleEditClick,
  } = useTenantManagement();

  // 加载选项数据
  const { loadTagOptions } = useTenantOptions();

  // 组件挂载时加载标签选项
  useEffect(() => {
    loadTagOptions();
  }, [loadTagOptions]);

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建租户成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新租户成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 状态管理成功后的回调
  const handleStatusManageSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdateStatus(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 租户操作处理（包含表格刷新）
  const handleTenantOperation = {
    delete: async (record: any) => {
      await handleDelete(record);
      handleTableRefresh();
    },
    toggleStatus: async (record: any) => {
      await handleToggleStatus(record);
      handleTableRefresh();
    },
    batchDelete: async (tenants: any[]) => {
      await handleBatchDelete(tenants);
      handleTableRefresh();
    },
    batchToggleStatus: async (
      tenants: any[],
      targetStatus: "active" | "inactive"
    ) => {
      await handleBatchToggleStatus(tenants, targetStatus);
      handleTableRefresh();
    },
  };

  return (
    <PageContainer subTitle="管理租户信息、配额设置和租户状态">
      {/* 主表格 */}
      <TenantTable
        actionRef={actionRef}
        selectionState={selectionState}
        onSelectionChange={setSelectionState}
        onCreateTenant={() => setModalOpen("create", true)}
        onEditTenant={handleEditClick}
        onDeleteTenant={handleTenantOperation.delete}
        onToggleStatus={handleTenantOperation.toggleStatus}
        onBatchDelete={handleTenantOperation.batchDelete}
        onBatchToggleStatus={handleTenantOperation.batchToggleStatus}
      />

      {/* 创建租户表单 */}
      <CreateTenantForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen("create", open)}
        onFinish={handleCreateSuccess}
      />

      {/* 编辑租户表单 */}
      <EditTenantForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />

      {/* 租户状态管理模态框 */}
      <StatusManageModal
        open={modalState.statusManage}
        onOpenChange={(open) => setModalOpen("statusManage", open)}
        onFinish={handleStatusManageSuccess}
        currentRow={currentRow}
      />
    </PageContainer>
  );
};

export default TenantManagement;
