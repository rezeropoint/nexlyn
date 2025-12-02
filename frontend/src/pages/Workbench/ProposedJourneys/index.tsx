/**
 * 我发起的流程页面
 * 展示用户发起的流程列表
 */
import { EyeOutlined } from "@ant-design/icons";
import { PageContainer, ProTable } from "@ant-design/pro-components";
import type { ActionType, ProColumns } from "@ant-design/pro-components";
import { Alert, Button, Select, Spin, Tag } from "antd";
import dayjs from "dayjs";
import React, { useEffect, useRef, useState } from "react";
import type { Journey } from "@/services/workbench/types";
import { getProposedJourneys } from "@/services/workbench";
import { getFlowList } from "@/services/eventhandler";
import type { FlowInfo } from "@/pages/EventManagement/types";
import FlowDetailDrawer from "../components/FlowDetailDrawer";
import { useApp } from "@/utils/appContext";
import styles from "./index.less";

/**
 * 我发起的流程页面组件
 */
const ProposedJourneys: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedFlowId, setSelectedFlowId] = useState<number>();
  const [selectedJourneyId, setSelectedJourneyId] = useState<number>();
  const { message } = useApp();

  // 流程列表相关状态
  const [flowList, setFlowList] = useState<FlowInfo[]>([]);
  const [flowListLoading, setFlowListLoading] = useState(false);
  const [currentFlowId, setCurrentFlowId] = useState<number>();

  /**
   * 加载流程列表
   */
  useEffect(() => {
    const fetchFlowList = async () => {
      setFlowListLoading(true);
      try {
        const res = await getFlowList(true); // 只获取已配置事件的流程
        if (res.code === 0 && res.data?.list) {
          setFlowList(res.data.list);
          // 默认选中第一个流程
          if (res.data.list.length > 0) {
            setCurrentFlowId(res.data.list[0].id);
          }
        } else {
          message.error(res.msg || "加载流程列表失败");
        }
      } catch (error) {
        message.error("加载流程列表失败");
        console.error("Failed to load flow list:", error);
      } finally {
        setFlowListLoading(false);
      }
    };

    fetchFlowList();
  }, [message]);

  /**
   * 查看详情
   */
  const handleViewDetail = (record: Journey) => {
    setSelectedFlowId(record.flowId);
    setSelectedJourneyId(record.id);
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
  const columns: ProColumns<Journey>[] = [
    {
      title: "流程编号",
      dataIndex: "sn",
      width: 180,
      ellipsis: true,
      copyable: true,
      render: (_, record) => (
        <div className={styles.flowCell}>
          <div className={styles.flowTitle}>{record.sn}</div>
          <div className={styles.flowMeta}>
            <span>流程ID {record.flowId}</span>
            <span>当前节点ID {record.currentVertexId ?? "-"}</span>
          </div>
        </div>
      ),
    },
    {
      title: "流程ID",
      dataIndex: "flowId",
      width: 100,
      search: false,
    },
    {
      title: "Journey ID",
      dataIndex: "id",
      width: 120,
      search: false,
    },
    {
      title: "当前节点ID",
      dataIndex: "currentVertexId",
      width: 120,
      search: false,
    },
    {
      title: "状态",
      dataIndex: "status",
      width: 120,
      render: (_, record) => {
        const statusMap: Record<
          Journey["status"],
          { text: string; className: string }
        > = {
          processing: { text: "进行中", className: styles.statusTagProcessing },
          completed: { text: "已完成", className: styles.statusTagCompleted },
          aborted: { text: "已终止", className: styles.statusTagError },
          stashed: { text: "已暂存", className: styles.statusTagDraft },
        };
        const current = statusMap[record.status];
        return (
          <Tag bordered={false} className={current?.className}>
            {current?.text || record.status}
          </Tag>
        );
      },
    },
    {
      title: "发起人",
      dataIndex: ["user", "name"],
      width: 120,
      search: false,
      render: (_, record) => {
        if (record.user) {
          return record.user.nickname || record.user.name;
        }
        return "-";
      },
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 180,
      search: false,
      render: (text) =>
        text ? dayjs(text as string).format("YYYY-MM-DD HH:mm:ss") : "-",
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 180,
      search: false,
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
    <PageContainer subTitle="追踪自己发起的每一个流程，快速了解流程状态">

      {/* 流程选择器 */}
      <div className={styles.flowSelector}>
        <Spin spinning={flowListLoading}>
          <Select
            showSearch
            allowClear
            placeholder="请选择或搜索流程"
            value={currentFlowId}
            onChange={(value) => {
              setCurrentFlowId(value);
              actionRef.current?.reload();
            }}
            options={flowList.map((flow) => ({
              label: flow.title,
              value: flow.id,
            }))}
            filterOption={(input, option) => {
              const title = option?.label as string;
              if (!title) return false;
              // 支持标题搜索（不区分大小写）
              return title.toLowerCase().includes(input.toLowerCase());
            }}
            optionFilterProp="label"
          />
        </Spin>
      </div>

      {/* 未选择流程时的提示 */}
      {!currentFlowId && !flowListLoading && (
        <Alert
          className={styles.emptyTip}
          message="请选择流程"
          description="请先从上方下拉框中选择要查看的流程类型，然后即可查看您在该流程下发起的所有记录。"
          type="info"
          showIcon
        />
      )}

      <ProTable<Journey>
        headerTitle="我发起的流程"
        actionRef={actionRef}
        columns={columns}
        rowKey="id"
        request={async (params) => {
          // 未选择流程时不加载数据
          if (!currentFlowId) {
            return {
              data: [],
              total: 0,
              success: true,
            };
          }

          try {
            const res = await getProposedJourneys({
              flowId: currentFlowId,
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
              message.error(res.msg || "加载流程列表失败");
              return {
                data: [],
                total: 0,
                success: false,
              };
            }
          } catch (error) {
            message.error("加载流程列表失败");
            console.error("Failed to load proposed journeys:", error);
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

export default ProposedJourneys;
