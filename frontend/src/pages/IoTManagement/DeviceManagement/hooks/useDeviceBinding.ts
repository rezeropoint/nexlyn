import {
  type BindDeviceRequest,
  type DeviceBinding,
  type DeviceBindingSummary,
  type UpdateDeviceRequest,
  bindDevice,
  getDevice,
  unbindDevice,
  updateDevice,
} from "@/services/iot";
import { Modal, message } from "antd";
import { useState } from "react";
import type { ModalState, SelectionState } from "../types";

/**
 * 设备管理业务逻辑Hook
 */
export function useDeviceBinding() {
  // 当前编辑的设备
  const [currentRow, setCurrentRow] = useState<
    DeviceBindingSummary | undefined
  >();

  // 当前查看详情的设备
  const [currentDevice, setCurrentDevice] = useState<
    DeviceBinding | undefined
  >();

  // 模态框状态
  const [modalState, setModalState] = useState<ModalState>({
    bind: false,
    edit: false,
    detail: false,
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
      if (type === "detail") {
        setCurrentDevice(undefined);
      }
    }
  };

  /**
   * 绑定设备
   */
  const handleBind = async (values: BindDeviceRequest): Promise<boolean> => {
    try {
      const response = await bindDevice(values);

      if (response.code === 0) {
        message.success("设备绑定成功");
        setModalOpen("bind", false);
        return true;
      } else {
        message.error(response.msg || "绑定失败");
        return false;
      }
    } catch (error: any) {
      message.error(`绑定失败: ${error.message || "未知错误"}`);
      return false;
    }
  };

  /**
   * 更新设备
   */
  const handleUpdate = async (
    values: UpdateDeviceRequest
  ): Promise<boolean> => {
    if (!currentRow) {
      message.error("未选择设备");
      return false;
    }

    try {
      // orgId 在 values 中（来自表单隐藏字段）
      const response = await updateDevice(
        currentRow.deviceId,
        values.orgId,
        values
      );

      if (response.code === 0) {
        message.success("设备更新成功");
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
   * 解绑设备
   */
  const handleUnbind = async (
    record: DeviceBindingSummary
  ): Promise<boolean> => {
    return new Promise((resolve) => {
      Modal.confirm({
        title: "确认解绑",
        content: `确定要解绑设备"${record.deviceName}"(${record.deviceId})吗？`,
        okText: "确定",
        cancelText: "取消",
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            const response = await unbindDevice(record.deviceId, record.orgId);

            if (response.code === 0) {
              message.success("设备解绑成功");
              resolve(true);
            } else {
              message.error(response.msg || "解绑失败");
              resolve(false);
            }
          } catch (error: any) {
            message.error(`解绑失败: ${error.message || "未知错误"}`);
            resolve(false);
          }
        },
        onCancel: () => {
          resolve(false);
        },
      });
    });
  };

  /**
   * 批量解绑设备
   */
  const handleBatchUnbind = async (
    devices: DeviceBindingSummary[]
  ): Promise<boolean> => {
    return new Promise((resolve) => {
      Modal.confirm({
        title: "确认批量解绑",
        content: `确定要解绑选中的 ${devices.length} 个设备吗？`,
        okText: "确定",
        cancelText: "取消",
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            let successCount = 0;
            let failCount = 0;

            for (const device of devices) {
              try {
                const response = await unbindDevice(
                  device.deviceId,
                  device.orgId
                );
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
              message.success(`成功解绑 ${successCount} 个设备`);
            }
            if (failCount > 0) {
              message.error(`${failCount} 个设备解绑失败`);
            }

            // 清空选择
            setSelectionState({
              selectedRowKeys: [],
              selectedRows: [],
            });

            resolve(successCount > 0);
          } catch (error: any) {
            message.error(`批量解绑失败: ${error.message || "未知错误"}`);
            resolve(false);
          }
        },
        onCancel: () => {
          resolve(false);
        },
      });
    });
  };

  /**
   * 查看设备详情
   */
  const handleViewDetail = async (record: DeviceBindingSummary) => {
    try {
      const response = await getDevice(record.deviceId, record.orgId);

      if (response.code === 0 && response.data) {
        setCurrentDevice(response.data);
        setModalOpen("detail", true);
      } else {
        message.error(response.msg || "获取设备详情失败");
      }
    } catch (error: any) {
      message.error(`获取设备详情失败: ${error.message || "未知错误"}`);
    }
  };

  /**
   * 处理编辑点击
   */
  const handleEditClick = (record: DeviceBindingSummary) => {
    setCurrentRow(record);
    setModalOpen("edit", true);
  };

  return {
    // 状态
    currentRow,
    currentDevice,
    modalState,
    selectionState,

    // 设置方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleBind,
    handleUpdate,
    handleUnbind,
    handleBatchUnbind,
    handleViewDetail,
    handleEditClick,
  };
}
