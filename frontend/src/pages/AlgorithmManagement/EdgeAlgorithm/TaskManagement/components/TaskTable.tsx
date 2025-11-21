import type { AIBoxAlgorithmTask, AIBoxCapabilities } from "@/services/iot";
import {
  PauseCircleFilled,
  PlayCircleFilled,
  PlayCircleOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { ProColumns } from "@ant-design/pro-components";
import { ProTable } from "@ant-design/pro-components";
import {
  Badge,
  Button,
  Empty,
  Popconfirm,
  Space,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import React, { useEffect, useMemo, useRef, useState } from "react";
import { TASK_STATUS_BADGE_MAP } from "../types";
import styles from "./TaskTable.less";

const { Text } = Typography;

interface TaskTableProps {
  tasks: AIBoxAlgorithmTask[]; // 任务列表数据
  loading?: boolean; // 加载状态
  deviceId?: string; // 当前设备ID
  isOnline?: boolean; // 设备在线状态
  capabilities?: AIBoxCapabilities | null; // 算法能力数据
  onRefresh?: () => void; // 刷新回调
  onControlTask?: (taskId: string, controlCommand: number) => void; // 控制任务回调
}

/**
 * AI Box 任务列表表格
 *
 * 功能：
 * - 展示选中AI Box设备的所有算法任务
 * - 任务会话ID、描述、媒体、算法、状态等信息
 * - 支持任务状态可视化（Badge）
 * - 支持任务启停控制
 * - 空状态友好提示
 */
const TaskTable: React.FC<TaskTableProps> = ({
  tasks,
  loading = false,
  deviceId,
  isOnline = false,
  capabilities,
  onRefresh,
  onControlTask,
}) => {
  // 按钮loading状态管理（记录正在操作的任务ID）
  const [loadingTaskId, setLoadingTaskId] = useState<string | null>(null);

  // 轮询相关ref
  const pollTimerRef = useRef<NodeJS.Timeout | null>(null);
  const pollCountRef = useRef(0);
  const originalStatusRef = useRef<string | undefined>(undefined);
  const controllingTaskIdRef = useRef<string | null>(null);

  /**
   * 性能优化：使用 useMemo 创建 item -> name 的映射表
   * 避免每次渲染时遍历整个 abilities 数组（O(n) -> O(1)）
   */
  const algorithmNameMap = useMemo(() => {
    if (!capabilities?.abilities) return new Map<number, string>();

    const map = new Map<number, string>();
    capabilities.abilities.forEach((ability) => {
      map.set(ability.item, ability.name);
    });
    return map;
  }, [capabilities]);

  /**
   * 清除轮询和loading状态
   */
  const clearPolling = () => {
    if (pollTimerRef.current) {
      clearTimeout(pollTimerRef.current);
      pollTimerRef.current = null;
    }
    pollCountRef.current = 0;
    originalStatusRef.current = undefined;
    controllingTaskIdRef.current = null;
    setLoadingTaskId(null);
  };

  /**
   * 启动轮询刷新
   */
  const startPolling = () => {
    if (pollCountRef.current >= 5) {
      // 达到最大次数（10秒），结束
      clearPolling();
      return;
    }

    pollCountRef.current++;
    pollTimerRef.current = setTimeout(() => {
      // 触发刷新
      onRefresh?.();
      // 继续下一次轮询
      startPolling();
    }, 2000);
  };

  /**
   * 监听tasks变化，检测状态是否更新
   */
  useEffect(() => {
    if (!controllingTaskIdRef.current || !originalStatusRef.current) {
      return;
    }

    const currentTask = tasks.find(
      (t) => t.alg_task_session === controllingTaskIdRef.current
    );
    const currentStatus = currentTask?.alg_task_status?.style;

    // 如果状态变化了，提前结束loading
    if (currentStatus && currentStatus !== originalStatusRef.current) {
      clearPolling();
    }
  }, [tasks]);

  /**
   * 组件卸载时清理
   */
  useEffect(() => {
    return () => {
      if (pollTimerRef.current) {
        clearTimeout(pollTimerRef.current);
      }
    };
  }, []);

  /**
   * 处理任务控制（带智能loading动画）
   */
  const handleControl = async (taskId: string, controlCommand: number) => {
    // 记录原始状态
    const currentTask = tasks.find((t) => t.alg_task_session === taskId);
    originalStatusRef.current = currentTask?.alg_task_status?.style;
    controllingTaskIdRef.current = taskId;

    setLoadingTaskId(taskId);

    // 调用父组件的控制方法
    if (onControlTask) {
      await onControlTask(taskId, controlCommand);
    }

    // 启动轮询检测状态变化（每2秒，最多5次=10秒）
    pollCountRef.current = 0;
    startPolling();
  };

  // 定义表格列
  const columns: ProColumns<AIBoxAlgorithmTask>[] = [
    {
      title: "任务名称",
      dataIndex: "alg_task_session",
      key: "alg_task_session",
      copyable: true,
    },
    {
      title: "任务描述",
      dataIndex: "task_desc",
      key: "task_desc",
      render: (text) => text || "-",
    },
    {
      title: "媒体名称",
      dataIndex: "media_name",
      key: "media_name",
      render: (text) => (
        <Space size={4}>
          <PlayCircleOutlined className={styles.mediaIcon} />
          <span>{text || "-"}</span>
        </Space>
      ),
    },
    {
      title: "算法",
      dataIndex: "user_data",
      key: "algorithm",
      render: (userData: any) => {
        const methodConfig = userData?.MethodConfig;
        return (
          <Space wrap size={[4, 4]}>
            {methodConfig && methodConfig.length > 0 ? (
              methodConfig.map((algId: number) => {
                const algName = algorithmNameMap.get(algId);
                return (
                  <Tag key={algId} color="blue" className={styles.algorithmTag}>
                    {algName || algId}
                  </Tag>
                );
              })
            ) : (
              <Text type="secondary">-</Text>
            )}
          </Space>
        );
      },
    },
    {
      title: "状态",
      dataIndex: "alg_task_status",
      key: "alg_task_status",
      render: (status: any) => {
        const statusConfig =
          TASK_STATUS_BADGE_MAP[status?.style] || TASK_STATUS_BADGE_MAP.normal;
        return (
          <Badge
            status={statusConfig.status}
            text={status?.label || statusConfig.text}
          />
        );
      },
    },
    {
      title: "上报地址",
      dataIndex: "metadata_url",
      key: "metadata_url",
      render: (text) =>
        text ? (
          <Text copyable>{text}</Text>
        ) : (
          <span className={styles.emptyText}>-</span>
        ),
    },
    {
      title: "操作",
      valueType: "option",
      key: "action",
      fixed: "right",
      width: 100,
      render: (_, record) => {
        // 判断任务是否正在运行（style为success表示运行中）
        const isRunning = record.alg_task_status?.style === "success";
        const isLoading = loadingTaskId === record.alg_task_session;
        const isDisabled = !isOnline || isLoading;

        return (
          <Space size="small">
            <Popconfirm
              title={isRunning ? "确定停止此任务？" : "确定启动此任务？"}
              onConfirm={() =>
                handleControl(record.alg_task_session, isRunning ? 0 : 1)
              }
              okText="确定"
              cancelText="取消"
              disabled={isDisabled}
            >
              <Tooltip title={!isOnline ? "设备离线，无法控制任务" : ""}>
                <Button
                  type="link"
                  size="small"
                  icon={
                    isRunning ? <PauseCircleFilled /> : <PlayCircleFilled />
                  }
                  danger={isRunning}
                  className={isRunning ? undefined : styles.startButton}
                  loading={isLoading}
                  disabled={isDisabled}
                >
                  {isRunning ? "停止" : "启动"}
                </Button>
              </Tooltip>
            </Popconfirm>
          </Space>
        );
      },
    },
  ];

  // 空状态渲染
  if (!deviceId) {
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="请先选择AI Box设备"
        className={styles.emptyContainer}
      />
    );
  }

  return (
    <ProTable<AIBoxAlgorithmTask>
      columns={columns}
      dataSource={tasks}
      loading={loading}
      rowKey="alg_task_session"
      search={false}
      dateFormatter="string"
      scroll={{ x: "max-content" }}
      toolBarRender={() => [
        deviceId && onRefresh && (
          <Button
            key="refresh"
            type="text"
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            title="刷新任务"
          />
        ),
      ]}
      options={{
        reload: false,
        setting: true,
        density: true,
      }}
      pagination={{
        defaultPageSize: 10,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条任务`,
      }}
      locale={{
        emptyText: loading ? "加载中..." : "暂无任务数据",
      }}
      cardProps={false}
      className={styles.tableContainer}
    />
  );
};

export default TaskTable;
