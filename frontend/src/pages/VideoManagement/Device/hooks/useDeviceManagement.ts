import type { DeviceTag, NexlynDevice } from "@/services/video";
import {
  bindDevice,
  getDeviceTagList,
  getUnboundDeviceList,
  unbindDevice,
  updateDeviceAlias,
  updateDeviceTags,
} from "@/services/video";
import { message } from "antd";
import { useCallback, useEffect, useState } from "react";
import type {
  DeviceAliasFormValues,
  DeviceBindFormValues,
  DeviceEditFormValues,
  DeviceModalState,
  DeviceSelectionState,
  DeviceTagsFormValues,
  UseDeviceManagementReturn,
} from "../types";
import { useDevicePermissions } from "./useDevicePermissions";

/**
 * 设备管理核心业务逻辑Hook
 * 集中处理设备管理相关的所有业务逻辑
 */
export const useDeviceManagement = (
  organizationId?: string
): UseDeviceManagementReturn => {
  const permissions = useDevicePermissions();

  // 状态管理
  const [currentDevice, setCurrentDevice] = useState<NexlynDevice>();
  const [modalState, setModalState] = useState<DeviceModalState>({
    bind: false,
    editAlias: false,
    editTags: false,
    edit: false,
  });
  const [selectionState, setSelectionStateInternal] =
    useState<DeviceSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });
  const [availableTags, setAvailableTags] = useState<DeviceTag[]>([]);
  const [unboundDevices, setUnboundDevices] = useState<NexlynDevice[]>([]);
  const [loadingUnbound, setLoadingUnbound] = useState(false);

  // 模态框控制
  const setModalOpen = useCallback(
    (modal: keyof DeviceModalState, open: boolean) => {
      setModalState((prev) => ({ ...prev, [modal]: open }));
    },
    []
  );

  // 选择状态控制
  const setSelectionState = useCallback(
    (state: Partial<DeviceSelectionState>) => {
      setSelectionStateInternal((prev) => ({ ...prev, ...state }));
    },
    []
  );

  // 刷新未绑定设备列表
  const refreshUnboundDevices = useCallback(async (): Promise<void> => {
    if (!permissions.canBind) return;
    // 组织ID未就绪时不请求，避免后端400
    if (!organizationId) {
      setUnboundDevices([]);
      return;
    }

    setLoadingUnbound(true);
    try {
      const res = await getUnboundDeviceList({
        page: 1,
        page_size: 100,
        organizationId,
      });
      if (res.code === 0) {
        setUnboundDevices(res.data.devices || []);
      } else {
        message.error(res.message);
        setUnboundDevices([]);
      }
    } catch (error) {
      console.error("获取未绑定设备列表失败:", error);
      // 网络错误等异常情况才显示前端错误信息
      message.error("网络请求失败，请检查网络连接");
      setUnboundDevices([]);
    } finally {
      setLoadingUnbound(false);
    }
  }, [permissions.canBind, organizationId]);

  // 刷新可用标签列表
  const refreshAvailableTags = useCallback(async (): Promise<void> => {
    try {
      const res = await getDeviceTagList({ page: 1, pageSize: 100 });
      if (res.code === 0) {
        setAvailableTags(res.data.tags || []);
      } else {
        console.warn("获取标签列表失败:", res.message);
        setAvailableTags([]);
      }
    } catch (error) {
      console.error("获取标签列表失败:", error);
      setAvailableTags([]);
    }
  }, []);

  // 绑定设备
  const handleBind = useCallback(
    async (values: DeviceBindFormValues): Promise<boolean> => {
      try {
        // 检查权限
        if (!permissions.canBind) {
          message.error("您没有绑定设备的权限");
          return false;
        }

        const res = await bindDevice(values);
        if (res.code === 0) {
          message.success(res.message);
          setModalOpen("bind", false);
          // 刷新未绑定设备列表
          await refreshUnboundDevices();
          return true;
        } else {
          message.error(res.message);
          return false;
        }
      } catch (error) {
        console.error("绑定设备失败:", error);
        message.error("网络请求失败，请检查网络连接");
        return false;
      }
    },
    [permissions.canBind, setModalOpen, refreshUnboundDevices]
  );

  // 解绑设备
  const handleUnbind = useCallback(
    async (device: NexlynDevice): Promise<void> => {
      try {
        // 检查权限
        if (!permissions.canUnbind) {
          message.error("您没有解绑设备的权限");
          return;
        }

        if (!device.deviceId) {
          message.error("设备ID不能为空");
          return;
        }
        const res = await unbindDevice(device.deviceId);
        if (res.code === 0) {
          message.success(res.message);
          // 如果解绑的是选中的设备，需要更新选中状态
          const newSelectedRowKeys = selectionState.selectedRowKeys.filter(
            (key) => key !== device.deviceId
          );
          const newSelectedRows = selectionState.selectedRows.filter(
            (row) => row.deviceId !== device.deviceId
          );
          setSelectionState({
            selectedRowKeys: newSelectedRowKeys,
            selectedRows: newSelectedRows,
          });
          // 刷新未绑定设备列表
          await refreshUnboundDevices();
        } else {
          message.error(res.message);
        }
      } catch (error) {
        console.error("解绑设备失败:", error);
        message.error("网络请求失败，请检查网络连接");
      }
    },
    [
      permissions.canUnbind,
      selectionState,
      setSelectionState,
      refreshUnboundDevices,
    ]
  );

  // 更新设备别名
  const handleUpdateAlias = useCallback(
    async (values: DeviceAliasFormValues): Promise<boolean> => {
      const deviceId = currentDevice?.deviceId || currentDevice?.device_id;
      if (!deviceId) return false;

      try {
        // 检查权限
        if (!permissions.canEdit) {
          message.error("您没有编辑设备的权限");
          return false;
        }

        const res = await updateDeviceAlias(deviceId, values);
        if (res.code === 0) {
          message.success(res.message);
          setCurrentDevice(undefined);
          setModalOpen("editAlias", false);
          return true;
        } else {
          message.error(res.message);
          return false;
        }
      } catch (error) {
        console.error("更新设备别名失败:", error);
        message.error("网络请求失败，请检查网络连接");
        return false;
      }
    },
    [currentDevice?.device_id, permissions.canEdit, setModalOpen]
  );

  // 更新设备标签
  const handleUpdateTags = useCallback(
    async (values: DeviceTagsFormValues): Promise<boolean> => {
      const deviceId = currentDevice?.deviceId || currentDevice?.device_id;
      if (!deviceId) return false;

      try {
        // 检查权限
        if (!permissions.canEdit) {
          message.error("您没有编辑设备的权限");
          return false;
        }

        const res = await updateDeviceTags(deviceId, values);
        if (res.code === 0) {
          message.success(res.message);
          setCurrentDevice(undefined);
          setModalOpen("editTags", false);
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
    [currentDevice?.device_id, permissions.canEdit, setModalOpen]
  );

  // 统一编辑设备（别名和标签）
  const handleEditDevice = useCallback(
    async (values: DeviceEditFormValues): Promise<boolean> => {
      if (!currentDevice?.deviceId) return false;

      try {
        // 检查权限
        if (!permissions.canEdit) {
          message.error("您没有编辑设备的权限");
          return false;
        }

        // 并发执行别名和标签更新
        const promises = [];

        // 更新别名
        promises.push(
          updateDeviceAlias(currentDevice.deviceId, {
            deviceAlias: values.deviceAlias,
          })
        );

        // 更新标签
        if (values.tags !== undefined) {
          promises.push(
            updateDeviceTags(currentDevice.deviceId, { tags: values.tags })
          );
        }

        const results = await Promise.all(promises);

        // 检查所有操作是否成功
        const allSuccess = results.every((res) => res.code === 0);

        if (allSuccess) {
          message.success("设备信息更新成功");
          setCurrentDevice(undefined);
          setModalOpen("edit", false);
          return true;
        } else {
          const failedResults = results.filter((res) => res.code !== 0);
          message.error(
            `更新失败: ${failedResults.map((res) => res.message).join(", ")}`
          );
          return false;
        }
      } catch (error) {
        console.error("更新设备信息失败:", error);
        message.error("网络请求失败，请检查网络连接");
        return false;
      }
    },
    [currentDevice?.deviceId, permissions.canEdit, setModalOpen]
  );

  // 点击统一编辑按钮
  const handleEditClick = useCallback(
    (device: NexlynDevice): void => {
      setCurrentDevice(device);
      setModalOpen("edit", true);
    },
    [setModalOpen]
  );

  // 点击编辑别名按钮
  const handleEditAliasClick = useCallback(
    (device: NexlynDevice): void => {
      setCurrentDevice(device);
      setModalOpen("editAlias", true);
    },
    [setModalOpen]
  );

  // 点击编辑标签按钮
  const handleEditTagsClick = useCallback(
    (device: NexlynDevice): void => {
      setCurrentDevice(device);
      setModalOpen("editTags", true);
    },
    [setModalOpen]
  );

  // 初始化时加载数据
  useEffect(() => {
    refreshAvailableTags();
    if (permissions.canBind && organizationId) {
      refreshUnboundDevices();
    }
  }, [
    permissions.canBind,
    organizationId,
    refreshAvailableTags,
    refreshUnboundDevices,
  ]);

  return {
    // 状态
    currentDevice,
    modalState,
    selectionState,
    availableTags,
    unboundDevices,
    loadingUnbound,

    // 操作方法
    setCurrentDevice,
    setModalOpen,
    setSelectionState,
    refreshUnboundDevices,
    refreshAvailableTags,

    // 业务操作
    handleBind,
    handleUnbind,
    handleUpdateAlias,
    handleUpdateTags,
    handleEditDevice,
    handleEditClick,
    handleEditAliasClick,
    handleEditTagsClick,
  };
};
