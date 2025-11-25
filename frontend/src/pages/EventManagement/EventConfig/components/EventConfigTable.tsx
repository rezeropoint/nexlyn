import type {
  EventConfigWithFields,
  ListEventConfigsRequest,
} from "@/pages/EventManagement/types";
import { getEventConfigList } from "@/services/eventhandler";
import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import type { ProColumns } from "@ant-design/pro-components";
import { type ActionType, ProTable } from "@ant-design/pro-components";
import { Button, Popconfirm, Tag } from "antd";
import React from "react";

interface EventConfigTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  onCreate: () => void;
  onEdit: (record: EventConfigWithFields) => void;
  onDelete: (record: EventConfigWithFields) => Promise<boolean>;
}

const EventConfigTable: React.FC<EventConfigTableProps> = ({
  actionRef,
  onCreate,
  onEdit,
  onDelete,
}) => {
  const columns: ProColumns<EventConfigWithFields>[] = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      search: false,
    },
    {
      title: "配置名称",
      dataIndex: "name",
      key: "name",
    },
    {
      title: "Skylark流程",
      dataIndex: "flowTitle",
      key: "flowTitle",
      search: false,
      render: (_, record) => (
        <Tag color="blue">{`${record.flowTitle} (ID: ${record.flowId})`}</Tag>
      ),
    },
    {
      title: "字段数量",
      dataIndex: "fieldsCount",
      key: "fieldsCount",
      search: false,
      render: (_, record) => record.fields?.length || 0,
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      key: "updatedAt",
      valueType: "dateTime",
      search: false,
    },
    {
      title: "操作",
      key: "action",
      search: false,
      render: (_, record) => (
        <>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => onEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除？"
            description="删除后无法恢复,请谨慎操作。"
            onConfirm={() => onDelete(record)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </>
      ),
    },
  ];

  return (
    <ProTable<EventConfigWithFields, ListEventConfigsRequest>
      headerTitle="事件配置列表"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: 120,
      }}
      toolBarRender={() => [
        <Button type="primary" key="primary" onClick={onCreate}>
          <PlusOutlined /> 新建
        </Button>,
      ]}
      request={async (params) => {
        const response = await getEventConfigList(params);
        return {
          data: response.data?.list || [],
          success: response.code === 0,
          total: response.total || 0,
        };
      }}
      columns={columns}
    />
  );
};

export default EventConfigTable;
