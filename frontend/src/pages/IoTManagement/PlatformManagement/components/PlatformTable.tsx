import type { PlatformMetadata } from "@/services/iot";
import * as iotApi from "@/services/iot";
import { PlusOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Button, Tag } from "antd";
import React from "react";
import { PLATFORM_TYPE_COLORS, PLATFORM_TYPE_NAMES } from "../constants";
import PlatformTableActions from "./PlatformTableActions";

interface PlatformTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  onCreatePlatform: () => void;
  onEditPlatform: (record: PlatformMetadata) => void;
  onDeletePlatform: (record: PlatformMetadata) => void;
}

/**
 * 平台配置列表表格组件
 */
const PlatformTable: React.FC<PlatformTableProps> = ({
  actionRef,
  onCreatePlatform,
  onEditPlatform,
  onDeletePlatform,
}) => {
  // 表格列定义
  const columns: ProColumns<PlatformMetadata>[] = [
    {
      title: "配置名称",
      dataIndex: "name",
      width: 200,
      ellipsis: true,
      search: false,
    },
    {
      title: "平台类型",
      dataIndex: "type",
      width: 140,
      search: false,
      render: (_, record) => (
        <Tag color={PLATFORM_TYPE_COLORS[record.type]}>
          {PLATFORM_TYPE_NAMES[record.type] || record.type}
        </Tag>
      ),
    },
    {
      title: "描述",
      dataIndex: "description",
      width: 250,
      ellipsis: true,
      search: false,
      render: (text) => text || "-",
    },
    {
      title: "启用状态",
      dataIndex: "enabled",
      width: 100,
      align: "center",
      search: false,
      render: (_, record) => (
        <Tag color={record.enabled ? "success" : "default"}>
          {record.enabled ? "已启用" : "已禁用"}
        </Tag>
      ),
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 180,
      valueType: "dateTime",
      search: false,
      sorter: true,
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 180,
      valueType: "dateTime",
      search: false,
      sorter: true,
    },
    {
      title: "操作",
      valueType: "option",
      width: 140,
      fixed: "right",
      render: (_, record) => (
        <PlatformTableActions
          record={record}
          onEdit={onEditPlatform}
          onDelete={onDeletePlatform}
        />
      ),
    },
  ];

  return (
    <ProTable<PlatformMetadata>
      headerTitle="平台配置管理"
      actionRef={actionRef}
      rowKey="id"
      columns={columns}
      scroll={{ x: 1200 }}
      search={false}
      request={async (params, _sort) => {
        try {
          const response = await iotApi.getPlatformList({
            page: params.current,
            pageSize: params.pageSize,
          });

          return {
            data: response.data?.list || [],
            success: response.code === 0,
            total: response.total || 0,
          };
        } catch (_error) {
          return {
            data: [],
            success: false,
            total: 0,
          };
        }
      }}
      pagination={{
        defaultPageSize: 10,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条记录`,
      }}
      options={{
        reload: true,
        density: true,
        setting: true,
      }}
      toolBarRender={() => [
        <Button
          type="primary"
          key="primary"
          icon={<PlusOutlined />}
          onClick={onCreatePlatform}
        >
          新建平台配置
        </Button>,
      ]}
      dateFormatter="string"
    />
  );
};

export default PlatformTable;
