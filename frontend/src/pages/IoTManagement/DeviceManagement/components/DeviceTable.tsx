import { type ListDevicesRequest, listDevices } from "@/services/iot";
import { formatDateTime } from "@/utils/date";
import { DisconnectOutlined, PlusOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Badge, Button, Modal, Space, Tag, message } from "antd";
import React from "react";
import {
  DEVICE_STATUS_MAP,
  DEVICE_STATUS_OPTIONS,
  ONLINE_STATUS_OPTIONS,
  TABLE_CONFIG,
} from "../constants";
import { useDeviceCategories } from "../hooks/useDeviceCategories";
import { useDeviceModels } from "../hooks/useDeviceModels";
import type { DeviceBindingSummary, SelectionState } from "../types";
import styles from "./DeviceTable.less";
import DeviceTableActions from "./DeviceTableActions";

interface DeviceTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectedOrgId: string;
  selectionState: SelectionState;
  onSelectionChange: (state: Partial<SelectionState>) => void;
  onBindDevice: () => void;
  onViewDevice: (record: DeviceBindingSummary) => void;
  onEditDevice: (record: DeviceBindingSummary) => void;
  onUnbindDevice: (record: DeviceBindingSummary) => void;
  onBatchUnbind: (devices: DeviceBindingSummary[]) => void;
}

/**
 * 设备表格组件
 */
const DeviceTable: React.FC<DeviceTableProps> = ({
  actionRef,
  selectedOrgId,
  selectionState,
  onSelectionChange,
  onBindDevice,
  onViewDevice,
  onEditDevice,
  onUnbindDevice,
  onBatchUnbind,
}) => {
  const { selectedRowKeys, selectedRows } = selectionState;

  // 加载设备类别和型号选项
  const { categoryOptions } = useDeviceCategories();
  const { modelOptions } = useDeviceModels();

  // 定义表格列
  const columns: ProColumns<DeviceBindingSummary>[] = [
    {
      title: "设备ID",
      dataIndex: "deviceId",
      ellipsis: {
        showTitle: true,
      },
      copyable: true,
      width: 120,
    },
    {
      title: "设备名称",
      dataIndex: "deviceName",
      ellipsis: true,
    },
    {
      title: "设备别名",
      dataIndex: "deviceAlias",
      ellipsis: true,
      search: false,
      render: (text) => text || "-",
    },
    {
      title: "设备型号",
      dataIndex: "deviceModel",
      ellipsis: true,
      valueType: "select",
      fieldProps: {
        options: modelOptions,
        showSearch: true,
        placeholder: "请选择设备型号",
      },
    },
    {
      title: "设备类别",
      dataIndex: "deviceCategory",
      ellipsis: true,
      valueType: "select",
      fieldProps: {
        options: categoryOptions,
        showSearch: true,
        placeholder: "请选择设备类别",
      },
    },
    {
      title: "在线状态",
      dataIndex: "isOnline",
      width: 90,
      valueType: "select",
      valueEnum: {
        true: { text: "在线", status: "Success" },
        false: { text: "离线", status: "Default" },
      },
      fieldProps: {
        options: ONLINE_STATUS_OPTIONS,
      },
      render: (_, record) =>
        record.isOnline ? (
          <Badge status="success" text="在线" />
        ) : (
          <Badge status="default" text="离线" />
        ),
    },
    {
      title: "设备状态",
      dataIndex: "status",
      width: 90,
      valueType: "select",
      valueEnum: Object.fromEntries(
        Object.entries(DEVICE_STATUS_MAP).map(([key, { text, color }]) => [
          key,
          { text, status: color },
        ])
      ),
      fieldProps: {
        options: DEVICE_STATUS_OPTIONS,
      },
      render: (_, record) => (
        <Tag
          color={
            DEVICE_STATUS_MAP[record.status as keyof typeof DEVICE_STATUS_MAP]
              ?.color
          }
        >
          {DEVICE_STATUS_MAP[record.status as keyof typeof DEVICE_STATUS_MAP]
            ?.text || record.status}
        </Tag>
      ),
    },
    {
      title: "安装位置",
      dataIndex: "location",
      ellipsis: true,
      search: false,
      render: (text) => text || "-",
    },
    {
      title: "标签",
      dataIndex: "tags",
      search: false,
      render: (_, record) => {
        const tags = record.tags;
        if (!tags || tags.length === 0) return "-";
        return (
          <Space size={[0, 4]} wrap>
            {tags.slice(0, 2).map((tag: any) => (
              <Tag key={tag.id} color={tag.color} className={styles.tagItem}>
                {tag.name}
              </Tag>
            ))}
            {tags.length > 2 && <Tag color="default">+{tags.length - 2}</Tag>}
          </Space>
        );
      },
    },
    {
      title: "最后数据时间",
      dataIndex: "lastDataAt",
      width: 160,
      valueType: "dateTime",
      search: false,
      render: (_, record) =>
        record.lastDataAt ? formatDateTime(record.lastDataAt) : "-",
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      width: 160,
      valueType: "dateTime",
      search: false,
      render: (_, record) => formatDateTime(record.createdAt),
    },
    {
      title: "操作",
      valueType: "option",
      width: 150,
      fixed: "right",
      render: (_, record) => (
        <DeviceTableActions
          record={record}
          onView={onViewDevice}
          onEdit={onEditDevice}
          onUnbind={onUnbindDevice}
        />
      ),
    },
  ];

  // 工具栏渲染
  const toolBarRender = () => {
    const hasSelected = selectedRowKeys.length > 0;

    return [
      // 批量解绑按钮
      <Button
        key="batchUnbind"
        danger
        icon={<DisconnectOutlined />}
        disabled={!hasSelected}
        onClick={() => {
          Modal.confirm({
            title: `确定要解绑选中的 ${selectedRows.length} 个设备吗？`,
            content: "解绑后设备将不再关联到组织",
            okText: "确定",
            cancelText: "取消",
            onOk: () => onBatchUnbind(selectedRows),
          });
        }}
      >
        批量解绑 {hasSelected && `(${selectedRows.length})`}
      </Button>,
      // 绑定设备按钮
      <Button
        key="bind"
        type="primary"
        icon={<PlusOutlined />}
        onClick={onBindDevice}
      >
        绑定设备
      </Button>,
    ];
  };

  return (
    <ProTable<DeviceBindingSummary, ListDevicesRequest>
      headerTitle="设备列表"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: 120,
      }}
      scroll={{ x: 1600 }}
      options={{
        reload: true,
        setting: true,
        density: true,
      }}
      pagination={{
        defaultPageSize: TABLE_CONFIG.DEFAULT_PAGE_SIZE,
        showQuickJumper: true,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 条`,
      }}
      toolBarRender={toolBarRender}
      request={async (params) => {
        // 如果没有选中组织，不发起请求
        if (!selectedOrgId) {
          return {
            data: [],
            success: true,
            total: 0,
          };
        }

        try {
          const response = await listDevices({
            page: params.current,
            pageSize: params.pageSize,
            orgId: selectedOrgId, // 使用选中的组织ID（必填）
            deviceModel: params.deviceModel,
            deviceCategory: params.deviceCategory,
            status: params.status,
            isOnline: params.isOnline,
            keyword: params.keyword,
          });

          if (response.code === 0 && response.data) {
            return {
              data: response.data.list || [],
              success: true,
              total: response.total || 0,
            };
          } else {
            message.error(
              response.msg || response.message || "获取设备列表失败"
            );
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        } catch (error) {
          console.error("获取设备列表失败:", error);
          return {
            data: [],
            success: false,
            total: 0,
          };
        }
      }}
      columns={columns}
      rowSelection={{
        selectedRowKeys,
        onChange: (selectedRowKeys, selectedRows) => {
          onSelectionChange({ selectedRowKeys, selectedRows });
        },
      }}
    />
  );
};

export default DeviceTable;
