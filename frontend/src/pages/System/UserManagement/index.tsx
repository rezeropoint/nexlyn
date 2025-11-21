import { type ActionType, PageContainer } from "@ant-design/pro-components";
import { App, Modal } from "antd";
import React, { useRef } from "react";
import { syncUser, unbindUser } from "@/services/sync";
import BindUserForm from "./components/BindUserForm";
import CreateUserForm from "./components/CreateUserForm";
import EditUserForm from "./components/EditUserForm";
import ResetPasswordModal from "./components/ResetPasswordModal";
import UserTable from "./components/UserTable";
import { useUserManagement } from "./hooks/useUserManagement";

/**
 * 用户管理主页面
 *
 * 重构后的组件结构：
 * - 主页面只负责组件编排和状态协调
 * - 具体业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const UserManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const { message } = App.useApp();

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentUser,
    currentRow,
    modalState,
    selectionState,

    // 操作方法
    setModalOpen,
    setSelectionState,
    setCurrentRow,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleResetPassword,
    handleBatchDelete,
    handleBatchToggleStatus,
    handleToggleStatus,
    handleEditClick,
  } = useUserManagement();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建用户成功后的回调
  const handleCreateSuccess = async (
    values: API.CreateUserRequest
  ): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新用户成功后的回调
  const handleUpdateSuccess = async (
    values: API.UpdateUserRequest
  ): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 重置密码成功后的回调
  const handleResetPasswordSuccess = async (values: {
    password: string;
  }): Promise<boolean> => {
    return await handleResetPassword(values);
  };

  // 用户操作处理（包含表格刷新）
  const handleUserOperation = {
    delete: async (record: API.UserBrief) => {
      await handleDelete(record);
      handleTableRefresh();
    },
    toggleStatus: async (record: API.UserBrief) => {
      await handleToggleStatus(record);
      handleTableRefresh();
    },
    batchDelete: async (users: API.UserBrief[]) => {
      await handleBatchDelete(users);
      handleTableRefresh();
    },
    batchToggleStatus: async (
      users: API.UserBrief[],
      targetStatus: "active" | "inactive"
    ) => {
      await handleBatchToggleStatus(users, targetStatus);
      handleTableRefresh();
    },
    resetPassword: (_record: API.UserBrief) => {
      setModalOpen("resetPassword", true);
      // currentRow 会在 handleStatusClick 中设置
    },
    syncUser: async (record: API.UserBrief) => {
      try {
        const response = await syncUser({ id: record.id });
        if (response.code === 0) {
          message.success(
            `用户 ${record.userName} 同步成功，远程用户ID: ${response.data?.remoteUserId}`
          );
          handleTableRefresh();
        } else {
          message.error(response.msg || "同步失败");
        }
      } catch (_error) {
        message.error("同步失败");
      }
    },
    bindUser: (record: API.UserBrief) => {
      setCurrentRow(record as any);
      setModalOpen("bind", true);
    },
    unbindUser: async (record: API.UserBrief) => {
      Modal.confirm({
        title: `确定要解除用户 ${record.userName} 的绑定吗？`,
        content: "解除绑定后将无法在 Skylark 中同步该用户，但不会删除远程用户数据",
        onOk: async () => {
          try {
            const response = await unbindUser({ id: record.id });
            if (response.code === 0) {
              message.success("解绑成功");
              handleTableRefresh();
            } else {
              message.error(response.msg || "解绑失败");
            }
          } catch (_error) {
            message.error("解绑失败");
          }
        },
      });
    },
  };

  return (
    <PageContainer subTitle="管理系统用户账号、角色分配和账号状态">
      {/* 主表格 */}
      <UserTable
        actionRef={actionRef}
        selectionState={selectionState}
        onSelectionChange={setSelectionState}
        onCreateUser={() => setModalOpen("create", true)}
        onEditUser={handleEditClick}
        onDeleteUser={handleUserOperation.delete}
        onToggleStatus={handleUserOperation.toggleStatus}
        onResetPassword={handleUserOperation.resetPassword}
        onBatchDelete={handleUserOperation.batchDelete}
        onBatchToggleStatus={handleUserOperation.batchToggleStatus}
        onSyncUser={handleUserOperation.syncUser}
        onBindUser={handleUserOperation.bindUser}
        onUnbindUser={handleUserOperation.unbindUser}
      />

      {/* 创建用户表单 */}
      <CreateUserForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen("create", open)}
        onFinish={handleCreateSuccess}
        currentUser={currentUser}
      />

      {/* 编辑用户表单 */}
      <EditUserForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />

      {/* 重置密码模态框 */}
      <ResetPasswordModal
        open={modalState.resetPassword}
        onOpenChange={(open) => setModalOpen("resetPassword", open)}
        onFinish={handleResetPasswordSuccess}
      />

      {/* 绑定用户表单 */}
      <BindUserForm
        open={modalState.bind}
        onOpenChange={(open) => setModalOpen("bind", open)}
        currentUser={currentRow}
        onSuccess={handleTableRefresh}
      />
    </PageContainer>
  );
};

export default UserManagement;
