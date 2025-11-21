import type {
  CreateTenantRequest,
  Tenant,
  UpdateTenantRequest,
  UpdateTenantStatusRequest,
} from "@/services/tenant";
import {
  createTenant,
  deleteTenant,
  updateTenant,
  updateTenantStatus,
} from "@/services/tenant";
import { useModel } from "@umijs/max";
import { message } from "antd";
import { useCallback, useState } from "react";
import type {
  ModalState,
  TenantSelectionState,
  UseTenantManagementReturn,
} from "../types";
import { useTenantPermissions } from "./useTenantPermissions";

/**
 * 租户管理核心业务逻辑Hook
 * 集中处理租户管理相关的所有业务逻辑
 */
export const useTenantManagement = (): UseTenantManagementReturn => {
  const { initialState } = useModel("@@initialState");
  const currentUser = initialState?.currentUser;
  const permissions = useTenantPermissions();

  // 状态管理
  const [currentRow, setCurrentRow] = useState<Tenant>();
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
    statusManage: false,
  });
  const [selectionState, setSelectionStateInternal] =
    useState<TenantSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });

  // 模态框控制
  const setModalOpen = useCallback((modal: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [modal]: open }));
  }, []);

  // 选择状态控制
  const setSelectionState = useCallback(
    (state: Partial<TenantSelectionState>) => {
      setSelectionStateInternal((prev) => ({ ...prev, ...state }));
    },
    []
  );

  // 创建租户
  const handleCreate = useCallback(
    async (values: CreateTenantRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canCreate) {
          message.error("您没有创建租户的权限");
          return false;
        }

        const res = await createTenant(values);

        if (res.code === 0) {
          message.success(res.msg || "创建租户成功");
          setModalOpen("create", false);
          return true;
        } else {
          message.error(res.msg || "创建失败");
          return false;
        }
      } catch (error) {
        console.error("创建租户失败:", error);
        message.error("创建租户失败");
        return false;
      }
    },
    [permissions.canCreate, setModalOpen]
  );

  // 更新租户
  const handleUpdate = useCallback(
    async (values: UpdateTenantRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有更新租户的权限");
          return false;
        }

        // 检查租户级别
        if (currentRow && !permissions.canOperateTenant(currentRow)) {
          message.error("您不能编辑该租户（权限不足）");
          return false;
        }

        if (!currentRow?.tenantKey) return false;

        const res = await updateTenant({ id: currentRow.id }, values);

        if (res.code === 0) {
          message.success(res.msg || "更新租户成功");
          setModalOpen("edit", false);
          setCurrentRow(undefined);
          return true;
        } else {
          message.error(res.msg || "更新失败");
          return false;
        }
      } catch (error) {
        console.error("更新租户失败:", error);
        message.error("更新租户失败");
        return false;
      }
    },
    [
      permissions.canUpdate,
      permissions.canOperateTenant,
      currentRow,
      setModalOpen,
    ]
  );

  // 删除租户
  const handleDelete = useCallback(
    async (record: Tenant): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除租户的权限");
          return;
        }

        // 检查租户级别
        if (!permissions.canOperateTenant(record)) {
          message.error("您不能删除该租户（权限不足）");
          return;
        }

        const res = await deleteTenant({ id: record.id });

        if (res.code === 0) {
          message.success(res.msg || "删除成功");
          // 如果删除的是选中的租户，需要更新选中状态
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
        console.error("删除租户失败:", error);
        message.error("删除失败");
      }
    },
    [
      permissions.canDelete,
      permissions.canOperateTenant,
      selectionState,
      setSelectionState,
    ]
  );

  // 更新租户状态
  const handleUpdateStatus = useCallback(
    async (values: UpdateTenantStatusRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有更新租户状态的权限");
          return false;
        }

        // 检查租户级别
        if (currentRow && !permissions.canOperateTenant(currentRow)) {
          message.error("您不能更新该租户的状态（权限不足）");
          return false;
        }

        if (!currentRow?.id) return false;

        const res = await updateTenantStatus({ id: currentRow.id }, values);

        if (res.code === 0) {
          message.success(res.msg || "租户状态更新成功");
          setModalOpen("statusManage", false);
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
      permissions.canOperateTenant,
      currentRow,
      setModalOpen,
    ]
  );

  // 批量删除租户
  const handleBatchDelete = useCallback(
    async (tenants: Tenant[]): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除租户的权限");
          return;
        }

        // 检查租户级别
        const unauthorizedTenants = tenants.filter(
          (tenant) => !permissions.canOperateTenant(tenant)
        );
        if (unauthorizedTenants.length > 0) {
          message.error(
            `您不能删除 ${unauthorizedTenants.length} 个租户（权限不足）`
          );
          return;
        }

        // 检查是否包含自己的租户
        const hasOwnTenant = tenants.some(
          (tenant) =>
            tenant.id === currentUser?.tenantInfo?.tenantId &&
            currentUser?.role !== "super_admin"
        );
        if (hasOwnTenant) {
          message.error("您不能删除自己所属的租户");
          return;
        }

        // 批量删除
        const deletePromises = tenants.map((tenant) =>
          deleteTenant({ id: tenant.id })
        );
        const results = await Promise.allSettled(deletePromises);

        let successCount = 0;
        let errorCount = 0;

        results.forEach((result, index) => {
          if (result.status === "fulfilled" && result.value.code === 0) {
            successCount++;
          } else {
            errorCount++;
            console.error(
              `删除租户 ${tenants[index].tenantName} 失败:`,
              result
            );
          }
        });

        if (successCount > 0) {
          message.success(`成功删除 ${successCount} 个租户`);
          setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        }

        if (errorCount > 0) {
          message.error(`${errorCount} 个租户删除失败`);
        }
      } catch (error) {
        console.error("批量删除租户失败:", error);
        message.error("批量删除失败");
      }
    },
    [
      permissions.canDelete,
      permissions.canOperateTenant,
      currentUser,
      setSelectionState,
    ]
  );

  // 批量切换租户状态
  const handleBatchToggleStatus = useCallback(
    async (
      tenants: Tenant[],
      targetStatus: "active" | "inactive"
    ): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有切换租户状态的权限");
          return;
        }

        // 检查租户级别
        const unauthorizedTenants = tenants.filter(
          (tenant) => !permissions.canOperateTenant(tenant)
        );
        if (unauthorizedTenants.length > 0) {
          message.error(
            `您不能操作 ${unauthorizedTenants.length} 个租户（权限不足）`
          );
          return;
        }

        // 批量更新状态
        const updatePromises = tenants.map((tenant) =>
          updateTenantStatus({ id: tenant.id }, { status: targetStatus })
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
              `更新租户 ${tenants[index].tenantName} 状态失败:`,
              result
            );
          }
        });

        if (successCount > 0) {
          message.success(
            `成功${
              targetStatus === "active" ? "启用" : "禁用"
            } ${successCount} 个租户`
          );
          setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        }

        if (errorCount > 0) {
          message.error(`${errorCount} 个租户状态更新失败`);
        }
      } catch (error) {
        console.error("批量更新租户状态失败:", error);
        message.error("批量更新状态失败");
      }
    },
    [permissions.canUpdate, permissions.canOperateTenant, setSelectionState]
  );

  // 快速切换租户状态
  const handleToggleStatus = useCallback(
    async (record: Tenant): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有切换租户状态的权限");
          return;
        }

        // 检查租户级别
        if (!permissions.canOperateTenant(record)) {
          message.error("您不能切换该租户的状态（权限不足）");
          return;
        }

        const newStatus = record.status === "active" ? "inactive" : "active";
        const res = await updateTenantStatus(
          { id: record.id },
          { status: newStatus }
        );

        if (res.code === 0) {
          message.success(`租户已${newStatus === "active" ? "启用" : "禁用"}`);
          // 更新选中状态中的租户状态
          const updatedSelectedRows = selectionState.selectedRows.map((row) =>
            row.id === record.id ? { ...row, status: newStatus as any } : row
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
      permissions.canOperateTenant,
      selectionState.selectedRows,
      setSelectionState,
    ]
  );

  // 点击编辑按钮
  const handleEditClick = useCallback(
    (record: Tenant): void => {
      setCurrentRow(record);
      setModalOpen("edit", true);
    },
    [setModalOpen]
  );

  // 点击状态管理按钮
  const handleStatusClick = useCallback(
    (record: Tenant): void => {
      setCurrentRow(record);
      setModalOpen("statusManage", true);
    },
    [setModalOpen]
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
    handleUpdateStatus,
    handleBatchDelete,
    handleBatchToggleStatus,
    handleToggleStatus,
    handleEditClick,
    handleStatusClick,
  };
};
