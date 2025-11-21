/**
 * 抄送消息页面
 * 展示用户收到的抄送消息列表
 */
import { EyeOutlined } from "@ant-design/icons";
import { PageContainer, ProTable } from "@ant-design/pro-components";
import type { ActionType, ProColumns } from "@ant-design/pro-components";
import { Button } from "antd";
import dayjs from "dayjs";
import React, { useRef, useState } from "react";
import type { Assignment } from "@/services/workbench/types";
import { getUserAssignments } from "@/services/workbench";
import FlowDetailDrawer from "../components/FlowDetailDrawer";
import { useApp } from "@/utils/appContext";
import styles from "./index.less";

/**
 * 抄送消息页面组件
 */
const CCMessages: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedFlowId, setSelectedFlowId] = useState<number>();
  const [selectedJourneyId, setSelectedJourneyId] = useState<number>();
  const { message } = useApp();

  /**
   * 查看详情
   */
  const handleViewDetail = (record: Assignment) => {
    if (!record.flowId || !record.journeyId) {
      message.warning("流程信息不完整，无法查看详情");
      return;
    }
    setSelectedFlowId(record.flowId);
    setSelectedJourneyId(record.journeyId);
    setDetailVisible(true);
  };

  /**
   * 关闭详情 Drawer
   */
  const handleCloseDetail = () => {
    setDetailVisible(false);
    setSelectedFlowId(undefined);
    setSelectedJourneyId(undefined);
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
            <span>流程ID {record.flowId}</span>
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
      title: "抄送时间",
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
    <PageContainer subTitle="快速查看需要知会我的审批进展，并了解流程详情">
      <ProTable<Assignment>
        headerTitle="抄送消息"
        actionRef={actionRef}
        columns={columns}
        rowKey="id"
        request={async (params) => {
          try {
            const res = await getUserAssignments({
              category: "cc",
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
              message.error(res.msg || "加载抄送消息失败");
              return {
                data: [],
                total: 0,
                success: false,
              };
            }
          } catch (error) {
            message.error("加载抄送消息失败");
            console.error("Failed to load CC messages:", error);
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
        onClose={handleCloseDetail}
      />
    </PageContainer>
  );
};

export default CCMessages;
