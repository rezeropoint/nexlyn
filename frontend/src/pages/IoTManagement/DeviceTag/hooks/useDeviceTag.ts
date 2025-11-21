import {
  type CreateTagRequest,
  type DeviceTagSummary,
  type UpdateTagRequest,
  createTag,
  deleteTag,
  updateTag,
} from "@/services/iot";
import { Modal, message } from "antd";
import { useState } from "react";
import type { ModalState, SelectionState } from "../types";

/**
 * 设备标签管理业务逻辑Hook
 */
export function useDeviceTag() {
  // 当前编辑的标签
  const [currentRow, setCurrentRow] = useState<DeviceTagSummary | undefined>();

  // 模态框状态
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
  });

  // 选择状态
  const [selectionState, setSelectionState] = useState<SelectionState>({
    selectedRowKeys: [],
    selectedRows: [],
  });

  /**
   * 设置模态框打开状态
   */
  const setModalOpen = (type: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [type]: open }));
    if (!open) {
      setCurrentRow(undefined);
    }
  };

  /**
   * 创建标签
   */
  const handleCreate = async (values: CreateTagRequest): Promise<boolean> => {
    try {
      const response = await createTag(values);

      if (response.code === 0) {
        message.success("标签创建成功");
        setModalOpen("create", false);
        return true;
      } else {
        message.error(response.msg || "创建失败");
        return false;
      }
    } catch (error: any) {
      message.error(`创建失败: ${error.message || "未知错误"}`);
      return false;
    }
  };

  /**
   * 更新标签
   */
  const handleUpdate = async (values: UpdateTagRequest): Promise<boolean> => {
    if (!currentRow) {
      message.error("未选择标签");
      return false;
    }

    try {
      const response = await updateTag(currentRow.id, values);

      if (response.code === 0) {
        message.success("标签更新成功");
        setModalOpen("edit", false);
        return true;
      } else {
        message.error(response.msg || "更新失败");
        return false;
      }
    } catch (error: any) {
      message.error(`更新失败: ${error.message || "未知错误"}`);
      return false;
    }
  };

  /**
   * 删除标签
   */
  const handleDelete = async (record: DeviceTagSummary) => {
    Modal.confirm({
      title: "确认删除",
      content: `确定要删除标签"${record.name}"吗？`,
      okText: "确定",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const response = await deleteTag(record.id);

          if (response.code === 0) {
            message.success("标签删除成功");
          } else {
            message.error(response.msg || "删除失败");
          }
        } catch (error: any) {
          message.error(`删除失败: ${error.message || "未知错误"}`);
        }
      },
    });
  };

  /**
   * 批量删除标签
   */
  const handleBatchDelete = async (tags: DeviceTagSummary[]) => {
    Modal.confirm({
      title: "确认批量删除",
      content: `确定要删除选中的 ${tags.length} 个标签吗？`,
      okText: "确定",
      cancelText: "取消",
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          let successCount = 0;
          let failCount = 0;

          for (const tag of tags) {
            try {
              const response = await deleteTag(tag.id);
              if (response.code === 0) {
                successCount++;
              } else {
                failCount++;
              }
            } catch {
              failCount++;
            }
          }

          if (successCount > 0) {
            message.success(`成功删除 ${successCount} 个标签`);
          }
          if (failCount > 0) {
            message.error(`${failCount} 个标签删除失败`);
          }

          // 清空选择
          setSelectionState({
            selectedRowKeys: [],
            selectedRows: [],
          });
        } catch (error: any) {
          message.error(`批量删除失败: ${error.message || "未知错误"}`);
        }
      },
    });
  };

  /**
   * 处理编辑点击
   */
  const handleEditClick = (record: DeviceTagSummary) => {
    setCurrentRow(record);
    setModalOpen("edit", true);
  };

  return {
    // 状态
    currentRow,
    modalState,
    selectionState,

    // 设置方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleBatchDelete,
    handleEditClick,
  };
}
