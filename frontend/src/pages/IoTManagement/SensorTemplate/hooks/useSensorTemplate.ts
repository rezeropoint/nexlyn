import {
  type CreateSensorTemplateRequest,
  type SensorTemplate,
  type UpdateSensorTemplateRequest,
  createSensorTemplate,
  deleteSensorTemplate,
  updateSensorTemplate,
} from "@/services/iot";
import { Modal, message } from "antd";
import { useState } from "react";

/**
 * 模态框状态类型
 */
interface ModalState {
  create: boolean;
  edit: boolean;
  detail: boolean;
}

/**
 * 设备模板管理业务逻辑Hook
 */
export const useSensorTemplate = () => {
  // 当前选中的行
  const [currentRow, setCurrentRow] = useState<SensorTemplate | undefined>();

  // 模态框状态
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
    detail: false,
  });

  /**
   * 设置模态框开关
   */
  const setModalOpen = (modalName: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [modalName]: open }));
    // 关闭模态框时清除当前行
    if (!open) {
      setCurrentRow(undefined);
    }
  };

  /**
   * 处理编辑点击
   */
  const handleEditClick = (record: SensorTemplate) => {
    setCurrentRow(record);
    setModalOpen("edit", true);
  };

  /**
   * 处理详情点击
   */
  const handleDetailClick = (record: SensorTemplate) => {
    setCurrentRow(record);
    setModalOpen("detail", true);
  };

  /**
   * 创建模板
   */
  const handleCreate = async (
    values: CreateSensorTemplateRequest
  ): Promise<boolean> => {
    const hide = message.loading("正在创建设备模板...");
    try {
      const response = await createSensorTemplate(values);
      hide();

      if (response.code === 0) {
        message.success("创建成功");
        setModalOpen("create", false);
        return true;
      } else {
        message.error(response.msg || "创建失败");
        return false;
      }
    } catch (error: any) {
      hide();
      message.error(error?.message || "创建失败，请重试");
      return false;
    }
  };

  /**
   * 更新模板
   */
  const handleUpdate = async (
    values: UpdateSensorTemplateRequest
  ): Promise<boolean> => {
    if (!currentRow?.id) {
      message.error("未选择要更新的模板");
      return false;
    }

    const hide = message.loading("正在更新设备模板...");
    try {
      const response = await updateSensorTemplate(currentRow.id, values);
      hide();

      if (response.code === 0) {
        message.success("更新成功");
        setModalOpen("edit", false);
        return true;
      } else {
        message.error(response.msg || "更新失败");
        return false;
      }
    } catch (error: any) {
      hide();
      message.error(error?.message || "更新失败，请重试");
      return false;
    }
  };

  /**
   * 删除模板
   */
  const handleDelete = async (record: SensorTemplate): Promise<void> => {
    Modal.confirm({
      title: "确认删除",
      content: `确定要删除设备模板 "${record.name}" (${record.model}) 吗？`,
      okText: "确认",
      cancelText: "取消",
      onOk: async () => {
        const hide = message.loading("正在删除...");
        try {
          const response = await deleteSensorTemplate(record.id);
          hide();

          if (response.code === 0) {
            message.success("删除成功");
            return Promise.resolve();
          } else {
            message.error(response.msg || "删除失败");
            return Promise.reject();
          }
        } catch (error: any) {
          hide();
          message.error(error?.message || "删除失败，请重试");
          return Promise.reject();
        }
      },
    });
  };

  /**
   * 切换启用状态
   */
  const handleToggleEnabled = async (record: SensorTemplate): Promise<void> => {
    const targetEnabled = !record.enabled;
    const hide = message.loading(
      `正在${targetEnabled ? "启用" : "禁用"}模板...`
    );

    try {
      const response = await updateSensorTemplate(record.id, {
        model: record.model,
        category: record.category,
        enabled: targetEnabled,
      });
      hide();

      if (response.code === 0) {
        message.success(`${targetEnabled ? "启用" : "禁用"}成功`);
      } else {
        message.error(response.msg || `${targetEnabled ? "启用" : "禁用"}失败`);
      }
    } catch (error: any) {
      hide();
      message.error(
        error?.message || `${targetEnabled ? "启用" : "禁用"}失败，请重试`
      );
    }
  };

  return {
    // 状态
    currentRow,
    modalState,

    // 设置方法
    setModalOpen,
    setCurrentRow,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleToggleEnabled,
    handleEditClick,
    handleDetailClick,
  };
};
