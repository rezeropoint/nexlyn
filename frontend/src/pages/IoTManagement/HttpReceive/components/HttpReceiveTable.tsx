import type { HttpReceiveMetadata } from "@/services/iot";
import * as iotApi from "@/services/iot";
import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Button, Popconfirm, Space, Switch, Tooltip, Typography } from "antd";
import React from "react";

interface HttpReceiveTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  onCreateConfig: () => void;
  onEditConfig: (record: HttpReceiveMetadata) => void;
  onDeleteConfig: (record: HttpReceiveMetadata) => void;
  onToggleEnabled: (record: HttpReceiveMetadata) => void;
}

/**
 * HTTP 接收配置列表表格组件
 */
const HttpReceiveTable: React.FC<HttpReceiveTableProps> = ({
  actionRef,
  onCreateConfig,
  onEditConfig,
  onDeleteConfig,
  onToggleEnabled,
}) => {

  // 表格列定义
  const columns: ProColumns<HttpReceiveMetadata>[] = [
    {
      title: "配置名称",
      dataIndex: "name",
      width: 180,
      ellipsis: true,
    },
    {
      title: "数据接收端点",
      dataIndex: "id",
      key: "endpoint",
      width: 300,
      search: false,
      render: (_, record) => (
        <Tooltip title="点击复制完整端点地址">
          <Typography.Text
            copyable={{ text: `/api/v1/data-receive/${record.id}` }}
            code
          >
            POST /api/v1/data-receive/{record.id.slice(0, 8)}...
          </Typography.Text>
        </Tooltip>
      ),
    },
    {
      title: "启用状态",
      dataIndex: "enabled",
      width: 100,
      align: "center",
      search: false,
      render: (_, record) => (
        <Switch
          checked={record.enabled}
          checkedChildren="启用"
          unCheckedChildren="禁用"
          onChange={() => onToggleEnabled(record)}
        />
      ),
    },
    {
      title: "描述",
      dataIndex: "description",
      width: 200,
      ellipsis: true,
      search: false,
      render: (text) => text || "-",
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
            onClick={() => onEditConfig(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除？"
            description="删除后无法恢复，请谨慎操作。"
            onConfirm={() => onDeleteConfig(record)}
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
    <ProTable<HttpReceiveMetadata>
      headerTitle="HTTP 接收配置"
      actionRef={actionRef}
      rowKey="id"
      columns={columns}
      scroll={{ x: 1500 }}
      search={{
        labelWidth: 120,
      }}
      request={async (params, _sort) => {
        try {
          const response = await iotApi.getHttpReceiveList({
            page: params.current,
            pageSize: params.pageSize,
            keyword: params.name,
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
        <Button type="primary" key="primary" onClick={onCreateConfig}>
          <PlusOutlined /> 新建配置
        </Button>,
      ]}
      dateFormatter="string"
    />
  );
};

export default HttpReceiveTable;
