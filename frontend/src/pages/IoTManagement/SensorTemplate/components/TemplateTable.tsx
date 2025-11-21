import { getSensorTemplateList, type SensorTemplate } from "@/services/iot";
import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Badge, Button, message, Popconfirm } from "antd";
import React from "react";
import { useDeviceCategories } from "../hooks/useDeviceCategories";

interface TemplateTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  onCreateTemplate: () => void;
  onEditTemplate: (record: SensorTemplate) => void;
  onDeleteTemplate: (record: SensorTemplate) => void;
}

/**
 * 设备模板表格组件
 */
const TemplateTable: React.FC<TemplateTableProps> = ({
  actionRef,
  onCreateTemplate,
  onEditTemplate,
  onDeleteTemplate,
}) => {
  // 加载设备类别选项
  const { deviceCategories } = useDeviceCategories(true);

  // 定义表格列
  const columns: ProColumns<SensorTemplate>[] = [
    {
      title: "模板名称",
      dataIndex: "name",
      ellipsis: true,
      formItemProps: {
        rules: [
          {
            required: true,
            message: "此项为必填项",
          },
        ],
      },
    },
    {
      title: "设备型号",
      dataIndex: "model",
      copyable: true,
      ellipsis: true,
      formItemProps: {
        rules: [
          {
            required: true,
            message: "此项为必填项",
          },
        ],
      },
    },
    {
      title: "设备类别",
      dataIndex: "category",
      width: 120,
      ellipsis: true,
      valueType: "select",
      fieldProps: {
        options: deviceCategories.map((cat) => ({
          label: cat.name,
          value: cat.code,
        })),
        showSearch: true,
        placeholder: "请选择设备类别",
      },
    },
    {
      title: "设备厂商",
      dataIndex: "manufacturer",
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: "版本",
      dataIndex: "version",
      width: 100,
      hideInSearch: true,
      render: (_, record) => record.version || "-",
    },
    {
      title: "启用状态",
      dataIndex: "enabled",
      width: 100,
      valueType: "select",
      valueEnum: {
        true: { text: "已启用", status: "Success" },
        false: { text: "已禁用", status: "Default" },
      },
      render: (_, record) =>
        record.enabled ? (
          <Badge status="success" text="已启用" />
        ) : (
          <Badge status="default" text="已禁用" />
        ),
    },
    {
      title: "设备数量",
      dataIndex: "deviceCount",
      width: 100,
      hideInSearch: true,
      sorter: true,
      render: (_, record) => record.deviceCount || 0,
    },
    {
      title: "描述",
      dataIndex: "description",
      ellipsis: true,
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 160,
      valueType: "dateTime",
      hideInSearch: true,
      sorter: true,
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      width: 160,
      valueType: "dateTime",
      hideInSearch: true,
      sorter: true,
    },
    {
      title: "操作",
      valueType: "option",
      width: 150,
      fixed: "right",
      render: (_, record) => [
        <Button
          key="edit"
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => onEditTemplate(record)}
        >
          编辑
        </Button>,
        <Popconfirm
          key="delete"
          title="确定要删除该模板吗？"
          description="删除后将无法恢复"
          onConfirm={() => onDeleteTemplate(record)}
          okText="确认"
          cancelText="取消"
        >
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <ProTable<SensorTemplate>
      headerTitle="设备模板列表"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: "auto",
      }}
      toolBarRender={() => [
        <Button
          key="create"
          type="primary"
          icon={<PlusOutlined />}
          onClick={onCreateTemplate}
        >
          新建模板
        </Button>,
      ]}
      request={async (params, _sort, _filter) => {
        const {
          current,
          pageSize,
          name,
          model,
          category,
          manufacturer,
          enabled,
        } = params;

        try {
          const response = await getSensorTemplateList({
            page: current,
            pageSize,
            name,
            model,
            category,
            manufacturer,
            enabled,
          });

          if (response.code === 0 && response.data) {
            return {
              data: response.data.list || [],
              success: true,
              total: response.total || 0,
            };
          } else {
            message.error(
              response.msg || response.message || "获取模板列表失败"
            );
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        } catch (_error) {
          return {
            data: [],
            success: false,
            total: 0,
          };
        }
      }}
      columns={columns}
      scroll={{ x: "max-content" }}
      pagination={{
        defaultPageSize: 10,
        showSizeChanger: true,
        showQuickJumper: true,
        showTotal: (total) => `共 ${total} 条`,
      }}
    />
  );
};

export default TemplateTable;
