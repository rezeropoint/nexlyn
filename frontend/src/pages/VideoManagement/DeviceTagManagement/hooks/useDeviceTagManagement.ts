import type { DeviceTag } from "@/services/video";
import {
  createDeviceTag,
  deleteDeviceTag,
  updateDeviceTag,
} from "@/services/video";
import { message } from "antd";
import { useCallback, useState } from "react";
import type {
  DeviceTagModalState,
  DeviceTagSelectionState,
  UseDeviceTagManagementReturn,
} from "../types";
import { useDeviceTagPermissions } from "./useDeviceTagPermissions";

/**
 * 设备标签管理核心业务逻辑Hook
 * 集中处理设备标签管理相关的所有业务逻辑
 */
export const useDeviceTagManagement = (): UseDeviceTagManagementReturn => {
  const permissions = useDeviceTagPermissions();

  // 状态管理
  const [currentRow, setCurrentRow] = useState<DeviceTag>();
  const [modalState, setModalState] = useState<DeviceTagModalState>({
    create: false,
    edit: false,
  });
  const [selectionState, setSelectionStateInternal] =
    useState<DeviceTagSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });

  // 模态框控制
  const setModalOpen = useCallback(
    (modal: keyof DeviceTagModalState, open: boolean) => {
      setModalState((prev) => ({ ...prev, [modal]: open }));
    },
    []
  );

  // 选择状态控制
  const setSelectionState = useCallback(
    (state: Partial<DeviceTagSelectionState>) => {
      setSelectionStateInternal((prev) => ({ ...prev, ...state }));
    },
    []
  );

  // 创建设备标签
  const handleCreate = useCallback(
    async (values: any): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canCreate) {
          message.error("您没有创建设备标签的权限");
          return false;
        }

        const res = await createDeviceTag(values);
        if (res.code === 0) {
          message.success(res.message);
          setModalOpen("create", false);
          return true;
        } else {
          message.error(res.message);
          return false;
        }
      } catch (error) {
        console.error("创建设备标签失败:", error);
        message.error("网络请求失败，请检查网络连接");
        return false;
      }
    },
    [permissions.canCreate, setModalOpen]
  );

  // 更新设备标签
  const handleUpdate = useCallback(
    async (values: any): Promise<boolean> => {
      if (!currentRow?.id) return false;

      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有更新设备标签的权限");
          return false;
        }

        const res = await updateDeviceTag(currentRow.id, values);
        if (res.code === 0) {
          message.success(res.message);
          setCurrentRow(undefined);
          setModalOpen("edit", false);
          return true;
        } else {
          message.error(res.message);
          return false;
        }
      } catch (error) {
        console.error("更新设备标签失败:", error);
        message.error("网络请求失败，请检查网络连接");
        return false;
      }
    },
    [currentRow?.id, permissions.canUpdate, setModalOpen]
  );

  // 删除设备标签
  const handleDelete = useCallback(
    async (record: DeviceTag): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除设备标签的权限");
          return;
        }

        const res = await deleteDeviceTag(record.id);
        if (res.code === 0) {
          message.success(res.message);
          // 如果删除的是选中的标签，需要更新选中状态
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
          message.error(res.message);
        }
      } catch (error) {
        console.error("删除设备标签失败:", error);
        message.error("网络请求失败，请检查网络连接");
      }
    },
    [permissions.canDelete, selectionState, setSelectionState]
  );

  // 批量删除设备标签
  const handleBatchDelete = useCallback(
    async (tags: DeviceTag[]): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除设备标签的权限");
          return;
        }

        // 批量删除
        const deletePromises = tags.map((tag) => deleteDeviceTag(tag.id));
        const results = await Promise.allSettled(deletePromises);

        let successCount = 0;
        let errorCount = 0;

        results.forEach((result, index) => {
          if (result.status === "fulfilled" && result.value.code === 0) {
            successCount++;
          } else {
            errorCount++;
            console.error(`删除设备标签 ${tags[index].tagName} 失败:`, result);
          }
        });

        if (successCount > 0) {
          message.success(`成功删除 ${successCount} 个设备标签`);
          setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        }

        if (errorCount > 0) {
          message.error(`${errorCount} 个设备标签删除失败`);
        }
      } catch (error) {
        console.error("批量删除设备标签失败:", error);
        message.error("网络请求失败，请检查网络连接");
      }
    },
    [permissions.canDelete, setSelectionState]
  );

  // 点击编辑按钮
  const handleEditClick = useCallback(
    (record: DeviceTag): void => {
      setCurrentRow(record);
      setModalOpen("edit", true);
    },
    [setModalOpen]
  );

  return {
    // 状态
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
    handleBatchDelete,
    handleEditClick,
  };
};
