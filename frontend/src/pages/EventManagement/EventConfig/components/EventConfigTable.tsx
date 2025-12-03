import type {
  EventConfigWithFields,
  ListEventConfigsRequest,
} from "@/pages/EventManagement/types";
import { getEventConfigList } from "@/services/eventhandler";
import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import type { ProColumns } from "@ant-design/pro-components";
import { type ActionType, ProTable } from "@ant-design/pro-components";
import { Button, Popconfirm, Space, Tag } from "antd";
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
      valueType: "option",
      width: 150,
      fixed: "right",
      render: (_, record) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => onEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除？"
            description="删除后无法恢复，请谨慎操作。"
            onConfirm={() => onDelete(record)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
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
      scroll={{ x: 1400 }}
      pagination={{
        defaultPageSize: 10,
        showQuickJumper: true,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 条`,
      }}
      options={{
        reload: true,
        setting: true,
        density: true,
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
