import type {
  ListOrgMappingsRequest,
  OrgMapping,
} from "@/pages/EventManagement/types";
import { getOrgMappingList } from "@/services/eventhandler";
import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import type { ProColumns } from "@ant-design/pro-components";
import { type ActionType, ProTable } from "@ant-design/pro-components";
import { Button, Popconfirm, Space, Tag } from "antd";
import React from "react";

interface OrgMappingTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  onCreate: () => void;
  onEdit: (record: OrgMapping) => void;
  onDelete: (record: OrgMapping) => Promise<boolean>;
}

const OrgMappingTable: React.FC<OrgMappingTableProps> = ({
  actionRef,
  onCreate,
  onEdit,
  onDelete,
}) => {
  const columns: ProColumns<OrgMapping>[] = [
    {
      title: "远程组织值",
      dataIndex: "remoteOrgValue",
      key: "remoteOrgValue",
      ellipsis: true,
      tooltip: "Skylark中的组织字段值",
    },
    {
      title: "本地组织",
      dataIndex: "localOrgName",
      key: "localOrgName",
      search: false,
      render: (_, record) => (
        <Tag color="cyan">{record.localOrgName || record.localOrgId}</Tag>
      ),
      tooltip: "映射到的本地组织",
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      key: "createdAt",
      valueType: "dateTime",
      search: false,
      width: 180,
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      key: "updatedAt",
      valueType: "dateTime",
      search: false,
      width: 180,
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
    <ProTable<OrgMapping, ListOrgMappingsRequest>
      headerTitle="组织映射列表"
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
        const response = await getOrgMappingList(params);
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

export default OrgMappingTable;
