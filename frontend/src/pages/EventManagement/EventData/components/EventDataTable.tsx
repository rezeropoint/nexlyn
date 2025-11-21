import { getEventDetail, queryEventData } from "@/services/eventhandler";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Select, Space } from "antd";
import { useApp } from "@/utils/appContext";
import React, { useEffect, useMemo, useRef, useState } from "react";
import type {
  ColumnInfo,
  EventConfigWithFields,
  EventDetailResponse,
} from "../../types";
import "./EventDataTable.less";
import EventDetailDrawer from "./EventDetailDrawer";

interface EventDataTableProps {
  eventConfig: EventConfigWithFields;
  selectedOrgId: string;
  selectedStatus?: string[]; // 状态筛选（可选，保留用于兼容）
}

const EventDataTable: React.FC<EventDataTableProps> = ({
  eventConfig,
  selectedOrgId,
  selectedStatus: externalSelectedStatus,
}) => {
  const { message } = useApp();
  const actionRef = useRef<ActionType>();
  const [dataSource, setDataSource] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [pagination, setPagination] = useState({ current: 1, pageSize: 10 });
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [detailData, setDetailData] = useState<
    EventDetailResponse["data"] | null
  >(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [dynamicColumns, setDynamicColumns] = useState<ProColumns<any>[]>([]);
  const [backendColumns, setBackendColumns] = useState<ColumnInfo[]>([]); // 保存后端返回的原始列信息
  const [selectedStatus, setSelectedStatus] = useState<string[]>(externalSelectedStatus || []); // 内部状态筛选

  // 动态生成表格列（优先使用后端返回的columns）
  const buildColumns = (backendColumns?: ColumnInfo[]): ProColumns<any>[] => {
    // 优先使用后端返回的列信息
    if (backendColumns && backendColumns.length > 0) {
      return backendColumns.map((col) => ({
        title: col.displayName || col.field,
        dataIndex: col.field,
        key: col.field,
        ellipsis: true,
        valueType:
          col.type === "datetime"
            ? "dateTime"
            : col.type === "number"
            ? "digit"
            : "text",
      }));
    }

    // 兜底：返回空数组，等待后端数据
    return [];
  };

  const tableColumns: ProColumns<any>[] = useMemo(() => {
    return [
      ...dynamicColumns,
      {
        title: "操作",
        key: "action",
        width: 120,
        fixed: "right",
        render: (_: any, record: any) => {
          // 使用 slp_journey_id 作为主键
          const journeyId = record.slp_journey_id;
          if (!journeyId) return null;
          return (
            <a
              className="action-link"
              onClick={() => handleViewDetail(journeyId)}
            >
              查看详情
            </a>
          );
        },
      },
    ];
  }, [dynamicColumns]);

  // 加载数据
  const loadData = async () => {
    if (!eventConfig.id || !selectedOrgId) return;

    setLoading(true);
    try {
      const response = await queryEventData(eventConfig.id, {
        eventId: eventConfig.id,
        orgId: selectedOrgId,
        page: pagination.current,
        pageSize: pagination.pageSize,
        status: selectedStatus, // 传递状态筛选参数
      });

      if (response.code === 0 && response.data) {
        // 使用后端返回的 records 字段，并为每条记录添加唯一索引
        const records = (response.data.records || []).map(
          (record: any, index: number) => ({
            ...record,
            _rowKey: `${pagination.current}_${index}`, // 添加唯一rowKey
          })
        );
        setDataSource(records);
        setTotal(response.data.total || 0);

        // 保存后端返回的原始列信息
        if (response.data.columns) {
          setBackendColumns(response.data.columns);
        }

        // 更新动态列（使用后端返回的 columns）
        const cols = buildColumns(response.data.columns);
        setDynamicColumns(cols);
      } else {
        message.error(response.msg || "查询失败");
      }
    } catch (_error) {
      message.error("查询失败");
    } finally {
      setLoading(false);
    }
  };

  // 查看详情
  const handleViewDetail = async (journeyId: number) => {
    setDetailLoading(true);
    setDrawerVisible(true);
    setDetailData(null); // 先清空旧数据
    try {
      const response = await getEventDetail(eventConfig.id, journeyId);

      if (response.code === 0 && response.data) {
        setDetailData(response.data);
      } else {
        message.error(response.msg || response.message || "获取详情失败");
      }
    } catch (error) {
      message.error("获取详情失败");
      console.error("获取事件详情失败:", error);
    } finally {
      setDetailLoading(false);
    }
  };

  // 分页变化
  const handlePaginationChange = (page: number, pageSize: number) => {
    setPagination({ current: page, pageSize });
  };

  // 依赖变化时重新加载
  useEffect(() => {
    loadData();
  }, [
    eventConfig.id,
    selectedOrgId,
    pagination.current,
    pagination.pageSize,
    selectedStatus,
  ]);

  return (
    <div className="event-data-table">
      <ProTable
        headerTitle="事件数据列表"
        actionRef={actionRef}
        rowKey="_rowKey"
        columns={tableColumns}
        dataSource={dataSource}
        loading={loading}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total,
          onChange: handlePaginationChange,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
        search={false}
        toolBarRender={() => [
          <Space key="toolbar" size="middle">
            <span style={{ fontSize: '14px', color: '#666' }}>状态筛选：</span>
            <Select
              mode="multiple"
              allowClear
              placeholder="请选择状态"
              style={{ minWidth: 280 }}
              value={selectedStatus}
              onChange={setSelectedStatus}
              maxTagCount="responsive"
              options={[
                // 虚拟状态
                { label: "待处理", value: "pending" },
                // 流程状态（processing既是流程状态也是虚拟状态）
                { label: "处理中", value: "processing" },
                { label: "已完成", value: "finished" },
                { label: "已终止", value: "aborted" },
              ]}
            />
          </Space>
        ]}
        scroll={{ x: "max-content" }}
        options={{
          reload: () => loadData(),
          density: true,
          fullScreen: true,
          setting: true,
        }}
      />

      <EventDetailDrawer
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        detail={detailData}
        loading={detailLoading}
        columns={backendColumns}
      />
    </div>
  );
};

export default EventDataTable;
