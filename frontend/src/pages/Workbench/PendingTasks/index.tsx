/**
 * 待办任务页面
 * 展示用户的待办任务列表
 */
import { EyeOutlined } from "@ant-design/icons";
import { PageContainer, ProTable } from "@ant-design/pro-components";
import type { ActionType, ProColumns } from "@ant-design/pro-components";
import { Button, Tag } from "antd";
import dayjs from "dayjs";
import React, { useEffect, useRef, useState } from "react";
import { history, useLocation } from "umi";
import type { Assignment } from "@/services/workbench/types";
import { getUserAssignments } from "@/services/workbench";
import FlowDetailDrawer from "../components/FlowDetailDrawer";
import { useApp } from "@/utils/appContext";
import styles from "./index.less";

/**
 * 待办任务页面组件
 */
const PendingTasks: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const location = useLocation();
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedFlowId, setSelectedFlowId] = useState<number>();
  const [selectedJourneyId, setSelectedJourneyId] = useState<number>();
  const [selectedAssignmentId, setSelectedAssignmentId] = useState<number>();
  const { message } = useApp();

  /**
   * 从 URL 参数中读取 flowId、journeyId 和 assignmentId，自动打开详情
   */
  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const flowId = params.get("flowId");
    const journeyId = params.get("journeyId");
    const assignmentId = params.get("assignmentId");

    if (flowId && journeyId) {
      setSelectedFlowId(Number(flowId));
      setSelectedJourneyId(Number(journeyId));
      // 如果URL参数中有assignmentId，也设置它
      if (assignmentId) {
        setSelectedAssignmentId(Number(assignmentId));
      }
      setDetailVisible(true);
    }
  }, [location.search]);

  /**
   * 查看详情
   */
  const handleViewDetail = (record: Assignment) => {
    if (!record.flowId || !record.journeyId || !record.id) {
      message.warning("流程信息不完整，无法查看详情");
      return;
    }
    setSelectedFlowId(record.flowId);
    setSelectedJourneyId(record.journeyId);
    setSelectedAssignmentId(record.id);
    setDetailVisible(true);
  };

  /**
   * 关闭详情 Drawer，并清除 URL 参数
   */
  const handleCloseDetail = () => {
    setDetailVisible(false);
    setSelectedFlowId(undefined);
    setSelectedJourneyId(undefined);
    setSelectedAssignmentId(undefined);

    // 清除 URL 参数，保持 URL 干净
    const params = new URLSearchParams(location.search);
    if (params.has("flowId") || params.has("journeyId") || params.has("assignmentId")) {
      history.replace("/workbench/pending-tasks");
    }
  };

  /**
   * 审批操作成功回调
   */
  const handleActionSuccess = () => {
    actionRef.current?.reload();
  };

  /**
   * 表格列定义
   */
  const columns: ProColumns<Assignment>[] = [
    {
      title: "流程标题",
      dataIndex: "flowTitle",
      ellipsis: true,
      width: 260,
      render: (_, record) => (
        <div className={styles.flowCell}>
          <div className={styles.flowTitle}>{record.flowTitle || "-"}</div>
          <div className={styles.flowMeta}>
            <span>流程ID {record.flowId ?? "-"}</span>
            <span>节点ID {record.vertexId ?? "-"}</span>
          </div>
        </div>
      ),
    },
    {
      title: "流程ID",
      dataIndex: "flowId",
      width: 100,
      hideInSearch: true,
    },
    {
      title: "Journey ID",
      dataIndex: "journeyId",
      width: 120,
      hideInSearch: true,
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 100,
      hideInSearch: true,
      render: (_, record) => (
        <Tag
          bordered={false}
          className={
            record.status === "processing"
              ? styles.statusTagProcessing
              : styles.statusTagCompleted
          }
        >
          {record.status === "processing" ? "处理中" : "已完成"}
        </Tag>
      ),
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 180,
      hideInSearch: true,
      render: (text) =>
        text ? dayjs(text as string).format("YYYY-MM-DD HH:mm:ss") : "-",
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 180,
      hideInSearch: true,
      render: (text) =>
        text ? dayjs(text as string).format("YYYY-MM-DD HH:mm:ss") : "-",
    },
    {
      title: "操作",
      valueType: "option",
      width: 120,
      fixed: "right",
      render: (_, record) => [
        <Button
          key="detail"
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        >
          查看详情
        </Button>,
      ],
    },
  ];

  return (
    <PageContainer subTitle="处理我负责的流程，支持查询、筛选与详情查看">
      <ProTable<Assignment>
        headerTitle="待办任务"
        actionRef={actionRef}
        columns={columns}
        rowKey="id"
        request={async (params) => {
          try {
            const res = await getUserAssignments({
              category: "processed",
              page: params.current || 1,
              pageSize: params.pageSize || 20,
            });

            if (res.code === 0) {
              return {
                data: res.data?.list || [],
                total: res.data?.total || 0,
                success: true,
              };
            } else {
              message.error(res.msg || "加载待办任务失败");
              return {
                data: [],
                total: 0,
                success: false,
              };
            }
          } catch (error) {
            message.error("加载待办任务失败");
            console.error("Failed to load pending tasks:", error);
            return {
              data: [],
              total: 0,
              success: false,
            };
          }
        }}
        pagination={{
          defaultPageSize: 20,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        search={false}
        dateFormatter="string"
        options={{
          reload: true,
          setting: true,
          density: true,
        }}
        scroll={{ x: "max-content" }}
      />

      <FlowDetailDrawer
        visible={detailVisible}
        flowId={selectedFlowId}
        journeyId={selectedJourneyId}
        assignmentId={selectedAssignmentId}
        onClose={handleCloseDetail}
        onActionSuccess={handleActionSuccess}
      />
    </PageContainer>
  );
};

export default PendingTasks;
