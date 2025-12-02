import { queryEventData } from "@/services/eventhandler";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Badge, message, Tag } from "antd";
import React, { useMemo, useRef, useState } from "react";
import {
  FLOW_STATUS_MAP,
  getProTableValueType,
  TABLE_CONFIG,
} from "../constants";
import type { EventConfigWithFields } from "../types";
import EventDetailDrawer from "./EventDetailDrawer";

interface EventDataTableProps {
  eventConfig: EventConfigWithFields;
  selectedOrgId: string; // 选中的组织ID
}

/**
 * 事件数据表格
 * 根据配置动态生成列和查询
 */
const EventDataTable: React.FC<EventDataTableProps> = ({
  eventConfig,
  selectedOrgId,
}) => {
  const actionRef = useRef<ActionType>();
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedJourneyId, setSelectedJourneyId] = useState<number | null>(
    null
  );

  // 根据字段配置动态生成ProTable列
  const columns: ProColumns[] = useMemo(() => {
    const baseColumns: ProColumns[] = [
      {
        title: "Journey ID",
        dataIndex: "slp_journey_id",
        width: 100,
        fixed: "left",
        search: false,
        render: (text) => <Tag>{text}</Tag>,
      },
      {
        title: "状态",
        dataIndex: "slp_status",
        width: 100,
        valueType: "select",
        valueEnum: Object.fromEntries(
          Object.entries(FLOW_STATUS_MAP).map(([key, { text, status }]) => [
            key,
            { text, status },
          ])
        ),
        render: (_, record) => {
          const status = record.slp_status as string;
          const config =
            FLOW_STATUS_MAP[status as keyof typeof FLOW_STATUS_MAP];
          return config ? (
            <Badge status={config.status as any} text={config.text} />
          ) : (
            <Tag>{status}</Tag>
          );
        },
      },
    ];

    // 动态字段列（根据配置生成）
    const dynamicColumns: ProColumns[] = eventConfig.fields
      .filter((field) => field.isVisible)
      .sort((a, b) => a.displayOrder - b.displayOrder)
      .map((field) => ({
        title: field.displayName,
        dataIndex: field.fieldName,
        width: 150,
        ellipsis: true,
        valueType: getProTableValueType(field.fieldType) as any,
        search: field.isSearchable,
        render: (text) => {
          // 处理不同类型的显示
          if (field.fieldType === "boolean") {
            return text ? (
              <Tag color="green">是</Tag>
            ) : (
              <Tag color="red">否</Tag>
            );
          }
          return text || "-";
        },
      }));

    // 系统字段（创建时间等）
    const systemColumns: ProColumns[] = [
      {
        title: "节点名称",
        dataIndex: "slp_vertex_name",
        width: 120,
        search: false,
        render: (text) => text || "-",
      },
      {
        title: "处理人",
        dataIndex: "slp_user_name",
        width: 100,
        search: false,
        render: (text) => text || "-",
      },
      {
        title: "创建时间",
        dataIndex: "slp_created_at",
        valueType: "dateTime",
        width: 160,
        search: false,
      },
      {
        title: "更新时间",
        dataIndex: "slp_updated_at",
        valueType: "dateTime",
        width: 160,
        search: false,
      },
    ];

    return [...baseColumns, ...dynamicColumns, ...systemColumns];
  }, [eventConfig]);

  // 处理行双击事件
  const handleRowDoubleClick = (record: any) => {
    setSelectedJourneyId(record.slp_journey_id);
    setDetailOpen(true);
  };

  return (
    <>
      <ProTable
        headerTitle={`${eventConfig.name} (${
          eventConfig.flowTitle || `流程${eventConfig.flowId}`
        })`}
        actionRef={actionRef}
        rowKey="slp_journey_id"
        search={{
          labelWidth: 120,
        }}
        request={async (params, sort) => {
          try {
            // 构建排序参数
            let sortField: string | undefined;
            let sortOrder: string | undefined;

            if (sort) {
              const sortKey = Object.keys(sort)[0];
              if (sortKey) {
                sortField = sortKey;
                sortOrder = sort[sortKey] === "ascend" ? "asc" : "desc";
              }
            }

            // 构建搜索参数（只传递有值的搜索字段）
            const searchParams: any = {
              orgId: selectedOrgId, // 必填的组织ID
              page: params.current,
              pageSize: params.pageSize,
              sortField,
              sortOrder,
            };

            // 收集搜索条件
            const searchFields: { [key: string]: any } = {};
            Object.keys(params).forEach((key) => {
              if (
                key !== "current" &&
                key !== "pageSize" &&
                key !== "_timestamp" &&
                params[key] !== undefined &&
                params[key] !== ""
              ) {
                searchFields[key] = params[key];
              }
            });

            // 如果有搜索条件，构建keyword（这里简化处理）
            if (Object.keys(searchFields).length > 0) {
              // 可以根据需要改进搜索逻辑
              searchParams.keyword = Object.values(searchFields)[0];
            }

            const response = await queryEventData(eventConfig.id, searchParams);

            if (response.code === 0 && response.data) {
              return {
                data: response.data.records || [],
                success: true,
                total: response.data.total || 0,
              };
            } else {
              message.error(response.msg || "获取事件数据失败");
              return {
                data: [],
                success: false,
                total: 0,
              };
            }
          } catch (error) {
            console.error("查询事件数据失败:", error);
            message.error("查询失败");
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        }}
        columns={columns}
        pagination={{
          defaultPageSize: TABLE_CONFIG.DEFAULT_PAGE_SIZE,
          showQuickJumper: true,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        scroll={{ x: 1600 }}
        onRow={(record) => ({
          onDoubleClick: () => handleRowDoubleClick(record),
          style: { cursor: "pointer" },
        })}
      />

      {selectedJourneyId && (
        <EventDetailDrawer
          open={detailOpen}
          onClose={() => {
            setDetailOpen(false);
            setSelectedJourneyId(null);
          }}
          eventConfigId={eventConfig.id}
          journeyId={selectedJourneyId}
        />
      )}
    </>
  );
};

export default EventDataTable;
