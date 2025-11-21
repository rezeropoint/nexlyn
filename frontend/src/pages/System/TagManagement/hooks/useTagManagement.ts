import { createTag, deleteTag, updateTag } from "@/services/tags";
import { message } from "antd";
import { useCallback, useState } from "react";
import { DEFAULTS } from "../constants";
import type {
  ModalState,
  TagSelectionState,
  UseTagManagementReturn,
} from "../types";
import { useTagPermissions } from "./useTagPermissions";

/**
 * 标签管理核心业务逻辑Hook
 * 集中处理标签管理相关的所有业务逻辑
 */
export const useTagManagement = (): UseTagManagementReturn => {
  const permissions = useTagPermissions();

  // 状态管理
  const [currentRow, setCurrentRow] = useState<API.TagDefinition>();
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
  });
  const [selectionState, setSelectionStateInternal] =
    useState<TagSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });

  // 默认作用域
  const defaultScope = DEFAULTS.SCOPE;

  // 模态框控制
  const setModalOpen = useCallback((modal: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [modal]: open }));
  }, []);

  // 选择状态控制
  const setSelectionState = useCallback((state: Partial<TagSelectionState>) => {
    setSelectionStateInternal((prev) => ({ ...prev, ...state }));
  }, []);

  // 创建标签
  const handleCreate = useCallback(
    async (values: API.CreateTagRequest): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canCreate) {
          message.error("您没有创建标签的权限");
          return false;
        }

        const res = await createTag(values);
        if (res.code === 0) {
          message.success(res.msg || "创建标签成功");
          setModalOpen("create", false);
          return true;
        } else {
          message.error(res.msg || "创建失败");
          return false;
        }
      } catch (_error) {
        message.error("创建标签失败");
        return false;
      }
    },
    [permissions.canCreate, setModalOpen]
  );

  // 更新标签
  const handleUpdate = useCallback(
    async (values: API.UpdateTagRequest): Promise<boolean> => {
      if (!currentRow?.id) return false;

      try {
        // 检查权限
        if (!permissions.canUpdate) {
          message.error("您没有更新标签的权限");
          return false;
        }

        const res = await updateTag({ id: currentRow.id }, values);
        if (res.code === 0) {
          message.success(res.msg || "更新标签成功");
          setCurrentRow(undefined);
          setModalOpen("edit", false);
          return true;
        } else {
          message.error(res.msg || "更新失败");
          return false;
        }
      } catch (_error) {
        message.error("更新标签失败");
        return false;
      }
    },
    [currentRow?.id, permissions.canUpdate, setModalOpen]
  );

  // 删除标签
  const handleDelete = useCallback(
    async (record: API.TagDefinition): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除标签的权限");
          return;
        }

        const res = await deleteTag({ id: record.id });
        if (res.code === 0) {
          message.success(res.msg || "删除标签成功");
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
          message.error(res.msg || "删除失败");
        }
      } catch (_error) {
        message.error("删除标签失败");
      }
    },
    [permissions.canDelete, selectionState, setSelectionState]
  );

  // 批量删除标签
  const handleBatchDelete = useCallback(
    async (tags: API.TagDefinition[]): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canDelete) {
          message.error("您没有删除标签的权限");
          return;
        }

        // 批量删除
        const deletePromises = tags.map((tag) => deleteTag({ id: tag.id }));
        const results = await Promise.allSettled(deletePromises);

        let successCount = 0;
        let errorCount = 0;

        results.forEach((result, index) => {
          if (result.status === "fulfilled" && result.value.code === 0) {
            successCount++;
          } else {
            errorCount++;
            console.error(`删除标签 ${tags[index].label} 失败:`, result);
          }
        });

        if (successCount > 0) {
          message.success(`成功删除 ${successCount} 个标签`);
          setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        }

        if (errorCount > 0) {
          message.error(`${errorCount} 个标签删除失败`);
        }
      } catch (error) {
        console.error("批量删除标签失败:", error);
        message.error("批量删除失败");
      }
    },
    [permissions.canDelete, setSelectionState]
  );

  // 点击编辑按钮
  const handleEditClick = useCallback(
    (record: API.TagDefinition): void => {
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
    defaultScope,

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
