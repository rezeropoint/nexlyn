import type { PlatformDetail, PlatformMetadata } from "@/services/iot";
import * as iotApi from "@/services/iot";
import { Modal, message } from "antd";
import { useState } from "react";

/**
 * 平台管理业务逻辑Hook
 *
 * 负责处理平台配置的CRUD操作和状态管理
 */

// 模态框状态类型
interface ModalState {
  create: boolean;
  edit: boolean;
}

export const usePlatform = () => {
  // === 状态管理 ===
  const [currentRow, setCurrentRow] = useState<PlatformDetail | undefined>(
    undefined
  );
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
  });

  // === 设置方法 ===

  /**
   * 设置模态框打开/关闭状态
   */
  const setModalOpen = (type: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [type]: open }));
  };

  // === 业务操作 ===

  /**
   * 创建平台配置
   */
  const handleCreate = async (values: any): Promise<boolean> => {
    try {
      const response = await iotApi.createPlatform(values);

      if (response.code === 0) {
        message.success("平台配置创建成功");
        setModalOpen("create", false);
        return true;
      } else {
        message.error(response.msg || "创建失败");
        return false;
      }
    } catch (error: any) {
      console.error("创建平台配置失败:", error);
      message.error(error.message || "创建失败，请稍后重试");
      return false;
    }
  };

  /**
   * 更新平台配置
   */
  const handleUpdate = async (values: any): Promise<boolean> => {
    if (!currentRow?.id) {
      message.error("未找到要更新的平台配置");
      return false;
    }

    try {
      const response = await iotApi.updatePlatform(currentRow.id, values);

      if (response.code === 0) {
        message.success("平台配置更新成功");
        setModalOpen("edit", false);
        setCurrentRow(undefined);
        return true;
      } else {
        message.error(response.msg || "更新失败");
        return false;
      }
    } catch (error: any) {
      console.error("更新平台配置失败:", error);
      message.error(error.message || "更新失败，请稍后重试");
      return false;
    }
  };

  /**
   * 删除单个平台配置
   */
  const handleDelete = async (record: PlatformMetadata): Promise<void> => {
    Modal.confirm({
      title: "确认删除",
      content: `确定要删除平台配置"${record.name}"吗？\n\n警告：删除后使用该平台配置的分发规则将失效！`,
      okText: "确认删除",
      okButtonProps: { danger: true },
      cancelText: "取消",
      onOk: async () => {
        try {
          const response = await iotApi.deletePlatform(record.id);

          if (response.code === 0) {
            message.success("平台配置删除成功");
          } else {
            message.error(response.msg || "删除失败");
          }
        } catch (error: any) {
          console.error("删除平台配置失败:", error);
          message.error(error.message || "删除失败，请稍后重试");
        }
      },
    });
  };

  /**
   * 处理编辑按钮点击
   */
  const handleEditClick = async (record: PlatformMetadata): Promise<void> => {
    try {
      // 获取平台配置完整信息（包含敏感配置）
      const response = await iotApi.getPlatform(record.id);

      if (response.code === 0 && response.data) {
        setCurrentRow(response.data);
        setModalOpen("edit", true);
      } else {
        message.error(response.msg || "获取平台配置详情失败");
      }
    } catch (error: any) {
      console.error("获取平台配置详情失败:", error);
      message.error(error.message || "获取平台配置详情失败，请稍后重试");
    }
  };

  /**
   * 切换启用/禁用状态
   */
  const handleToggleEnabled = async (
    record: PlatformMetadata
  ): Promise<boolean> => {
    const newStatus = !record.enabled;
    const actionText = newStatus ? "启用" : "禁用";

    try {
      const response = await iotApi.updatePlatform(record.id, {
        type: record.type,
        enabled: newStatus,
      });

      if (response.code === 0) {
        message.success(`${actionText}成功`);
        return true;
      } else {
        message.error(response.msg || `${actionText}失败`);
        return false;
      }
    } catch (error: any) {
      console.error(`${actionText}平台配置失败:`, error);
      message.error(error.message || `${actionText}失败，请稍后重试`);
      return false;
    }
  };

  return {
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
  };
};
