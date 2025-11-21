import type { AIBoxAlgorithmTask, AIBoxCapabilities } from "@/services/iot";
import {
  controlAIBoxTask,
  getAIBoxCapabilities,
  listAIBoxTasks,
  listDevices,
} from "@/services/iot";
import { useApp } from "@/utils/appContext";
import { useCallback, useState } from "react";
import type { AIBoxDeviceOption } from "../types";
import { convertDeviceToOption } from "../types";

/**
 * AI Box设备控制业务逻辑Hook
 */
export function useAIBoxControl() {
  const { message } = useApp();
  // ===== 状态管理 =====

  // 组织选择
  const [selectedOrgId, setSelectedOrgId] = useState<string>("");

  // 设备列表和选择
  const [deviceOptions, setDeviceOptions] = useState<AIBoxDeviceOption[]>([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string | undefined>(
    undefined
  );
  const [deviceLoading, setDeviceLoading] = useState(false);

  // 任务列表
  const [tasks, setTasks] = useState<AIBoxAlgorithmTask[]>([]);
  const [tasksLoading, setTasksLoading] = useState(false);

  // 算法能力
  const [capabilities, setCapabilities] = useState<AIBoxCapabilities | null>(
    null
  );
  const [capabilitiesLoading, setCapabilitiesLoading] = useState(false);

  // Drawer状态
  const [drawerOpen, setDrawerOpen] = useState(false);

  // ===== 业务方法 =====

  /**
   * 加载任务列表
   */
  const loadTasks = useCallback(async (deviceId: string, orgId: string) => {
    if (!deviceId || !orgId) {
      setTasks([]);
      return;
    }

    try {
      setTasksLoading(true);

      const response = await listAIBoxTasks(deviceId, orgId);

      if (response.code === 0 && response.data?.list) {
        setTasks(response.data.list);
      } else {
        message.error(response.msg || "加载任务列表失败");
        setTasks([]);
      }
    } catch (error) {
      console.error("加载任务列表失败:", error);
      message.error("加载任务列表失败");
      setTasks([]);
    } finally {
      setTasksLoading(false);
    }
  }, []);

  /**
   * 加载AI Box设备列表
   */
  const loadAIBoxDevices = useCallback(
    async (orgId: string) => {
      if (!orgId) {
        setDeviceOptions([]);
        setSelectedDeviceId(undefined);
        return;
      }

      try {
        setDeviceLoading(true);

        // 查询该组织下所有AI Box设备（category='ai_box'）
        const response = await listDevices({
          orgId,
          deviceCategory: "ai_box",
          page: 1,
          pageSize: 1000, // 一次性加载所有AI Box设备
        });

        if (response.code === 0 && response.data?.list) {
          const options = response.data.list.map(convertDeviceToOption);
          setDeviceOptions(options);

          // 如果有设备且当前没有选中设备，自动选中第一个在线设备或第一个设备
          if (options.length > 0 && !selectedDeviceId) {
            const onlineDevice = options.find((d) => d.isOnline);
            const autoSelectedDeviceId =
              onlineDevice?.deviceId || options[0].deviceId;
            setSelectedDeviceId(autoSelectedDeviceId);

            // 自动加载任务列表
            await loadTasks(autoSelectedDeviceId, orgId);
          }
        } else {
          message.error(response.msg || "加载AI Box设备列表失败");
          setDeviceOptions([]);
        }
      } catch (error) {
        console.error("加载AI Box设备失败:", error);
        message.error("加载AI Box设备列表失败");
        setDeviceOptions([]);
      } finally {
        setDeviceLoading(false);
      }
    },
    [selectedDeviceId, loadTasks]
  );

  /**
   * 加载算法能力
   */
  const loadCapabilities = useCallback(
    async (deviceId: string, orgId: string) => {
      if (!deviceId || !orgId) {
        setCapabilities(null);
        return;
      }

      try {
        setCapabilitiesLoading(true);

        const response = await getAIBoxCapabilities(deviceId, orgId);

        if (response.code === 0 && response.data) {
          setCapabilities(response.data);
        } else {
          message.error(response.msg || "加载算法能力失败");
          setCapabilities(null);
        }
      } catch (error) {
        console.error("加载算法能力失败:", error);
        message.error("加载算法能力失败");
        setCapabilities(null);
      } finally {
        setCapabilitiesLoading(false);
      }
    },
    []
  );

  /**
   * 处理组织选择
   */
  const handleOrgSelect = useCallback(
    async (orgId: string) => {
      setSelectedOrgId(orgId);
      setSelectedDeviceId(undefined);
      setTasks([]);
      setCapabilities(null);

      // 自动加载该组织下的AI Box设备
      await loadAIBoxDevices(orgId);
    },
    [loadAIBoxDevices]
  );

  /**
   * 处理设备选择
   */
  const handleDeviceChange = useCallback(
    async (deviceId: string) => {
      setSelectedDeviceId(deviceId);
      setCapabilities(null);

      // 自动加载该设备的任务列表
      if (deviceId && selectedOrgId) {
        await loadTasks(deviceId, selectedOrgId);
      }
    },
    [selectedOrgId, loadTasks]
  );

  /**
   * 打开算法能力抽屉
   */
  const handleOpenCapabilities = useCallback(async () => {
    if (!selectedDeviceId || !selectedOrgId) {
      message.warning("请先选择设备");
      return;
    }

    // 检查设备是否在线
    const selectedDevice = deviceOptions.find(
      (d) => d.deviceId === selectedDeviceId
    );
    if (!selectedDevice?.isOnline) {
      message.warning("设备离线，无法查看算法能力");
      return;
    }

    setDrawerOpen(true);

    // 加载算法能力数据
    await loadCapabilities(selectedDeviceId, selectedOrgId);
  }, [selectedDeviceId, selectedOrgId, deviceOptions, loadCapabilities]);

  /**
   * 关闭算法能力抽屉
   */
  const handleCloseCapabilities = useCallback(() => {
    setDrawerOpen(false);
  }, []);

  /**
   * 刷新任务列表
   */
  const refreshTasks = useCallback(async () => {
    if (selectedDeviceId && selectedOrgId) {
      await loadTasks(selectedDeviceId, selectedOrgId);
    }
  }, [selectedDeviceId, selectedOrgId, loadTasks]);

  /**
   * 控制任务（启动/停止）
   */
  const handleControlTask = useCallback(
    async (taskId: string, controlCommand: number) => {
      if (!selectedDeviceId || !selectedOrgId) {
        message.warning("请先选择设备");
        return;
      }

      // 检查设备是否在线
      const selectedDevice = deviceOptions.find(
        (d) => d.deviceId === selectedDeviceId
      );
      if (!selectedDevice?.isOnline) {
        message.warning("设备离线，无法控制任务");
        return;
      }

      const actionText = controlCommand === 1 ? "启动" : "停止";
      const loadingMessage = message.loading(`正在${actionText}任务...`, 0);

      try {
        const response = await controlAIBoxTask(
          selectedDeviceId,
          taskId,
          selectedOrgId,
          controlCommand
        );

        loadingMessage();

        if (response.code === 0) {
          message.success(
            controlCommand === 1
              ? "启动命令已发送，设备正在处理中，状态稍后更新"
              : "停止命令已发送成功"
          );

          // 立即刷新任务列表
          await refreshTasks();
        } else {
          message.error(response.msg || `${actionText}任务失败`);
        }
      } catch (error) {
        loadingMessage();
        console.error(`控制任务失败:`, error);
        message.error(`${actionText}任务失败`);
      }
    },
    [selectedDeviceId, selectedOrgId, deviceOptions, refreshTasks]
  );

  return {
    // 状态
    selectedOrgId,
    deviceOptions,
    selectedDeviceId,
    deviceLoading,
    tasks,
    tasksLoading,
    capabilities,
    capabilitiesLoading,
    drawerOpen,

    // 方法
    handleOrgSelect,
    handleDeviceChange,
    handleOpenCapabilities,
    handleCloseCapabilities,
    refreshTasks,
    handleControlTask,
  };
}
