import {
  createUser,
  deleteUser,
  getCurrentUser,
  getUserById,
  resetUserPassword,
  updateUser,
  updateUserStatus,
} from "@/services/user";
import { message } from "antd";
import { useCallback, useEffect, useState } from "react";
// API 类型通过全局声明获取
import type {
  ModalState,
  UserSelectionState,
  UseUserManagementReturn,
} from "../types";
import { useUserPermissions } from "./useUserPermissions";

/**
 * 用户管理核心业务逻辑Hook
 * 集中处理用户管理相关的所有业务逻辑
 */
export const useUserManagement = (): UseUserManagementReturn => {
  const permissions = useUserPermissions();

  // 状态管理
  const [currentUser, setCurrentUser] = useState<API.User>();
  const [currentRow, setCurrentRow] = useState<API.User>();
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
    resetPassword: false,
    bind: false,
  });
  const [selectionState, setSelectionStateInternal] =
    useState<UserSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });

  // 加载当前用户信息
  const loadCurrentUser = useCallback(async () => {
    try {
      const res = await getCurrentUser();
      if (res.code === 0 && res.data) {
        setCurrentUser(res.data);
      }
    } catch (error) {
      console.error("获取当前用户信息失败:", error);
    }
  }, []);

  // 组件挂载时加载当前用户信息
  useEffect(() => {
    loadCurrentUser();
  }, [loadCurrentUser]);

  // 模态框控制
  const setModalOpen = useCallback((modal: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [modal]: open }));
  }, []);

  // 选择状态控制
  const setSelectionState = useCallback(
    (state: Partial<UserSelectionState>) => {
      setSelectionStateInternal((prev) => ({ ...prev, ...state }));
    },
    []
  );

  // 创建用户
  const handleCreate = useCallback(
    async (values: API.CreateUserRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canCreate) {
          message.error("您没有创建用户的权限");
          return false;
        }

        // 如果不是超级管理员，自动设置为当前用户的租户
        if (!currentUser?.roles?.includes("super_admin")) {
          values.tenantId = currentUser?.tenantInfo?.tenantId || "";
        }

        const res = await createUser(values);

        if (res.code === 0) {
          message.success(res.msg || "创建成功");
          setModalOpen("create", false);
          return true;
        } else {
          message.error(res.msg || "创建失败");
          return false;
        }
      } catch (error) {
        console.error("创建用户失败:", error);
        message.error("创建失败");
        return false;
      }
    },
    [permissions.canCreate, currentUser, setModalOpen]
  );

  // 更新用户
  const handleUpdate = useCallback(
    async (values: API.UpdateUserRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有更新用户的权限");
          return false;
        }

        // 检查用户级别
        if (currentRow && !permissions.canOperateUser(currentRow as any)) {
          message.error("您不能编辑该用户（权限级别不足）");
          return false;
        }

        if (!currentRow?.id) return false;

        const res = await updateUser({ id: currentRow.id }, values);

        if (res.code === 0) {
          message.success(res.msg || "更新成功");
          setModalOpen("edit", false);
          return true;
        } else {
          message.error(res.msg || "更新失败");
          return false;
        }
      } catch (error) {
        console.error("更新用户失败:", error);
        message.error("更新失败");
        return false;
      }
    },
    [
      permissions.canUpdate,
      permissions.canOperateUser,
      currentRow,
      setModalOpen,
    ]
  );

  // 删除用户
  const handleDelete = useCallback(
    async (record: API.UserBrief): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除用户的权限");
          return;
        }

        // 检查用户级别
        if (!permissions.canOperateUser(record)) {
          message.error("您不能删除该用户（权限级别不足）");
          return;
        }

        if (!record.id) return;

        const res = await deleteUser({ id: record.id });

        if (res.code === 0) {
          message.success(res.msg || "删除成功");
          // 如果删除的是选中的用户，需要更新选中状态
          const newSelectedRowKeys = selectionState.selectedRowKeys.filter(
            (key) => key !== record.id
          );
          const newSelectedRows = selectionState.selectedRows.filter(
            (row) => row.id !== record.id
          );
          setSelectionState({
            selectedRowKeys: newSelectedRowKeys,
            selectedRows: newSelectedRows,
          });
        } else {
          message.error(res.msg || "删除失败");
        }
      } catch (error) {
        console.error("删除用户失败:", error);
        message.error("删除失败");
      }
    },
    [
      permissions.canDelete,
      permissions.canOperateUser,
      selectionState,
      setSelectionState,
    ]
  );

  // 重置密码
  const handleResetPassword = useCallback(
    async (values: { password: string }): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有重置密码的权限");
          return false;
        }

        // 检查用户级别
        if (currentRow && !permissions.canOperateUser(currentRow as any)) {
          message.error("您不能重置该用户的密码（权限级别不足）");
          return false;
        }

        if (!currentRow?.id) return false;

        const res = await resetUserPassword(
          { id: currentRow.id },
          { password: values.password }
        );

        if (res.code === 0) {
          message.success(res.msg || "密码重置成功");
          setModalOpen("resetPassword", false);
          return true;
        } else {
          message.error(res.msg || "重置密码失败");
          return false;
        }
      } catch (error) {
        console.error("重置密码失败:", error);
        message.error("重置密码失败");
        return false;
      }
    },
    [
      permissions.canUpdate,
      permissions.canOperateUser,
      currentRow,
      setModalOpen,
    ]
  );

  // 更新用户状态
  const handleUpdateStatus = useCallback(
    async (values: API.UpdateUserStatusRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有更新用户状态的权限");
          return false;
        }

        // 检查用户级别
        if (currentRow && !permissions.canOperateUser(currentRow as any)) {
          message.error("您不能更新该用户的状态（权限级别不足）");
          return false;
        }

        if (!currentRow?.id) return false;

        const res = await updateUserStatus({ id: currentRow.id }, values);

        if (res.code === 0) {
          message.success(res.msg || "用户状态更新成功");
          return true;
        } else {
          message.error(res.msg || "更新状态失败");
          return false;
        }
      } catch (error) {
        console.error("更新状态失败:", error);
        message.error("更新状态失败");
        return false;
      }
    },
    [
      permissions.canUpdate,
      permissions.canOperateUser,
      currentRow,
      setModalOpen,
    ]
  );

  // 批量删除用户
  const handleBatchDelete = useCallback(
    async (users: API.UserBrief[]): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除用户的权限");
          return;
        }

        // 检查用户级别
        const unauthorizedUsers = users.filter(
          (user) => !permissions.canOperateUser(user)
        );
        if (unauthorizedUsers.length > 0) {
          message.error(
            `您不能删除 ${unauthorizedUsers.length} 个用户（权限级别不足）`
          );
          return;
        }

        // 批量删除
        const deletePromises = users
          .filter((user) => user.id)
          .map((user) => deleteUser({ id: user.id as string }));
        const results = await Promise.allSettled(deletePromises);

        let successCount = 0;
        let errorCount = 0;

        results.forEach((result, index) => {
          if (result.status === "fulfilled" && result.value.code === 0) {
            successCount++;
          } else {
            errorCount++;
            console.error(`删除用户 ${users[index].userName} 失败:`, result);
          }
        });

        if (successCount > 0) {
          message.success(`成功删除 ${successCount} 个用户`);
          setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        }

        if (errorCount > 0) {
          message.error(`${errorCount} 个用户删除失败`);
        }
      } catch (error) {
        console.error("批量删除用户失败:", error);
        message.error("批量删除失败");
      }
    },
    [permissions.canDelete, permissions.canOperateUser, setSelectionState]
  );

  // 批量切换用户状态
  const handleBatchToggleStatus = useCallback(
    async (
      users: API.UserBrief[],
      targetStatus: "active" | "inactive"
    ): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有切换用户状态的权限");
          return;
        }

        // 检查用户级别
        const unauthorizedUsers = users.filter(
          (user) => !permissions.canOperateUser(user)
        );
        if (unauthorizedUsers.length > 0) {
          message.error(
            `您不能操作 ${unauthorizedUsers.length} 个用户（权限级别不足）`
          );
          return;
        }

        // 批量更新状态
        const updatePromises = users
          .filter((user) => user.id)
          .map((user) =>
            updateUserStatus(
              { id: user.id as string },
              { status: targetStatus }
            )
          );
        const results = await Promise.allSettled(updatePromises);

        let successCount = 0;
        let errorCount = 0;

        results.forEach((result, index) => {
          if (result.status === "fulfilled" && result.value.code === 0) {
            successCount++;
          } else {
            errorCount++;
            console.error(
              `更新用户 ${users[index].userName} 状态失败:`,
              result
            );
          }
        });

        if (successCount > 0) {
          message.success(
            `成功${
              targetStatus === "active" ? "启用" : "禁用"
            } ${successCount} 个用户`
          );
          setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        }

        if (errorCount > 0) {
          message.error(`${errorCount} 个用户状态更新失败`);
        }
      } catch (error) {
        console.error("批量更新用户状态失败:", error);
        message.error("批量更新状态失败");
      }
    },
    [permissions.canUpdate, permissions.canOperateUser, setSelectionState]
  );

  // 快速切换用户状态
  const handleToggleStatus = useCallback(
    async (record: API.UserBrief): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有切换用户状态的权限");
          return;
        }

        // 检查用户级别
        if (!permissions.canOperateUser(record)) {
          message.error("您不能切换该用户的状态（权限级别不足）");
          return;
        }

        if (!record.id) return;

        const newStatus = record.status === "active" ? "inactive" : "active";
        const res = await updateUserStatus(
          { id: record.id },
          { status: newStatus }
        );

        if (res.code === 0) {
          message.success(
            res.msg || `用户已${newStatus === "active" ? "启用" : "禁用"}`
          );
          // 更新选中状态中的用户状态
          const updatedSelectedRows = selectionState.selectedRows.map((row) =>
            row.id === record.id ? { ...row, status: newStatus } : row
          );
          setSelectionState({ selectedRows: updatedSelectedRows });
        } else {
          message.error(res.msg || "更新状态失败");
        }
      } catch (error) {
        console.error("更新状态失败:", error);
        message.error("更新状态失败");
      }
    },
    [
      permissions.canUpdate,
      permissions.canOperateUser,
      selectionState.selectedRows,
      setSelectionState,
    ]
  );

  // 点击编辑按钮
  const handleEditClick = useCallback(
    async (record: API.UserBrief): Promise<void> => {
      try {
        // 检查用户级别
        if (!permissions.canOperateUser(record)) {
          message.error("您不能编辑该用户（权限级别不足）");
          return;
        }

        if (!record.id) return;

        const res = await getUserById({ id: record.id });

        if (res.code === 0 && res.data) {
          setCurrentRow(res.data);
          setModalOpen("edit", true);
        } else {
          message.error(res.msg || "获取用户详情失败");
        }
      } catch (error) {
        console.error("获取用户详情失败:", error);
        message.error("获取用户详情失败");
      }
    },
    [permissions.canOperateUser, setModalOpen]
  );

  return {
    // 状态
    currentUser,
    currentRow,
    modalState,
    selectionState,

    // 操作方法
    setCurrentRow,
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleResetPassword,
    handleUpdateStatus,
    handleBatchDelete,
    handleBatchToggleStatus,
    handleToggleStatus,
    handleEditClick,
  };
};
