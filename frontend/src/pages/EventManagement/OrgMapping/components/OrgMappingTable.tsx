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
      title: "ID",
      dataIndex: "id",
      key: "id",
      search: false,
      width: 280,
      ellipsis: true,
    },
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
      key: "action",
      search: false,
      width: 150,
      fixed: "right",
      render: (_, record) => (
        <Space>
          <a onClick={() => onEdit(record)}>
            <EditOutlined /> 编辑
          </a>
          <Popconfirm
            title="确认删除？"
            description="删除后无法恢复，请谨慎操作。"
            onConfirm={() => onDelete(record)}
            okText="确认"
            cancelText="取消"
          >
            <a>
              <DeleteOutlined /> 删除
            </a>
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
