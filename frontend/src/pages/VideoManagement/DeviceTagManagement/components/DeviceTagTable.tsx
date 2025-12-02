import type { DeviceTag } from "@/services/video";
import { getDeviceTagList } from "@/services/video";
import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Button, message, Modal, Popconfirm, Tag } from "antd";
import dayjs from "dayjs";
import React from "react";
import { useDeviceTagPermissions } from "../hooks/useDeviceTagPermissions";
import type { DeviceTagSelectionState } from "../types";
import styles from "./DeviceTagTable.less";

// 表格配置
const DEVICE_TAG_TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 20,
  SEARCH_LABEL_WIDTH: 120,
  SCROLL_X: 1000,
} as const;

interface DeviceTagTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: DeviceTagSelectionState;
  onSelectionChange: (state: Partial<DeviceTagSelectionState>) => void;
  onCreateTag: () => void;
  onEditTag: (record: DeviceTag) => void;
  onDeleteTag: (record: DeviceTag) => void;
  onBatchDelete: (tags: DeviceTag[]) => void;
}

/**
 * 设备标签表格组件
 */
const DeviceTagTable: React.FC<DeviceTagTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateTag,
  onEditTag,
  onDeleteTag,
  onBatchDelete,
}) => {
  const permissions = useDeviceTagPermissions();
  const { selectedRowKeys, selectedRows } = selectionState;

  // 表格列定义
  const columns: ProColumns<DeviceTag>[] = [
    {
      title: "标签ID",
      dataIndex: "id",
      ellipsis: true,
      copyable: true,
      search: false,
      width: 180,
    },
    {
      title: "标签名称",
      dataIndex: "tagName",
      ellipsis: true,
      width: 150,
      render: (_, record) => (
        <Tag color="blue" style={{ fontWeight: "bold" }}>
          {record.tagName}
        </Tag>
      ),
    },
    {
      title: "描述",
      dataIndex: "description",
      ellipsis: true,
      search: false,
      width: 200,
    },
    {
      title: "使用数量",
      dataIndex: "deviceCount",
      search: false,
      width: 100,
      render: (_, record) => (
        <Tag color={record.deviceCount > 0 ? "green" : "default"}>
          {record.deviceCount || 0} 个设备
        </Tag>
      ),
    },
    {
      title: "创建时间",
      dataIndex: "createdAt",
      valueType: "dateTime",
      search: false,
      sorter: true,
      width: 180,
      render: (_, record) =>
        dayjs(record.createdAt).format("YYYY-MM-DD HH:mm:ss"),
    },
    {
      title: "更新时间",
      dataIndex: "updatedAt",
      valueType: "dateTime",
      search: false,
      sorter: true,
      width: 180,
      render: (_, record) =>
        dayjs(record.updatedAt).format("YYYY-MM-DD HH:mm:ss"),
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      fixed: "right",
      width: 150,
      render: (_, record) =>
        [
          permissions.canUpdate && (
            <Button
              key="edit"
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => onEditTag(record)}
            >
              编辑
            </Button>
          ),
          permissions.canDelete && (
            <Popconfirm
              key="delete"
              title="确定要删除这个设备标签吗？"
              description={`删除标签可能会影响 ${
                record.deviceCount || 0
              } 个设备，请谨慎操作`}
              onConfirm={() => onDeleteTag(record)}
              okText="确定"
              cancelText="取消"
              disabled={record.deviceCount > 0}
            >
              <Button
                type="link"
                size="small"
                danger
                icon={<DeleteOutlined />}
                disabled={record.deviceCount > 0}
                title={
                  record.deviceCount > 0
                    ? "有设备使用此标签，无法删除"
                    : "删除标签"
                }
              >
                删除
              </Button>
            </Popconfirm>
          ),
        ].filter(Boolean),
    },
  ];

  // 权限检查
  if (!permissions.canRead) {
    return (
      <div className={styles.permissionDenied}>
        <h2>权限不足</h2>
        <p>您需要设备标签管理权限才能访问此页面</p>
      </div>
    );
  }

  return (
    <ProTable<DeviceTag>
      headerTitle="设备标签管理"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: DEVICE_TAG_TABLE_CONFIG.SEARCH_LABEL_WIDTH,
        defaultCollapsed: false,
      }}
      toolBarRender={() => {
        const hasSelected = selectedRowKeys.length > 0;
        const canBatchDelete =
          hasSelected &&
          permissions.canDelete &&
          selectedRows.every((row) => (row.deviceCount || 0) === 0);

        return [
          // 多选操作按钮
          hasSelected && canBatchDelete && (
            <Button
              key="batch-delete"
              danger
              onClick={() => {
                Modal.confirm({
                  title: `确定要删除选中的 ${selectedRows.length} 个设备标签吗？`,
                  content: "删除标签后相关设备将失去这些标签",
                  okText: "确定",
                  cancelText: "取消",
                  onOk: () => onBatchDelete(selectedRows),
                });
              }}
              icon={<DeleteOutlined />}
            >
              批量删除
            </Button>
          ),
          // 创建标签按钮
          permissions.canCreate && (
            <Button
              key="create-tag"
              type="primary"
              onClick={onCreateTag}
              icon={<PlusOutlined />}
            >
              创建设备标签
            </Button>
          ),
        ].filter(Boolean);
      }}
      request={async (params) => {
        try {
          const res = await getDeviceTagList({
            page: params.current,
            pageSize: params.pageSize,
            keyword: params.tagName || params.keyword,
          });

          return {
            data: res.data?.tags || [],
            success: res.code === 0,
            total: res.data?.total || 0,
          };
        } catch (error) {
          console.error("获取设备标签列表失败:", error);
          message.error("获取设备标签列表失败");
          return { data: [], success: false, total: 0 };
        }
      }}
      columns={columns}
      scroll={{ x: DEVICE_TAG_TABLE_CONFIG.SCROLL_X }}
      rowSelection={{
        selectedRowKeys,
        onChange: (selectedRowKeys, selectedRows) => {
          onSelectionChange({ selectedRowKeys, selectedRows });
        },
        onSelect: (record, selected, _selectedRows) => {
          console.log("选中/取消选中设备标签:", record.tagName, selected);
        },
        onSelectAll: (selected, selectedRows, _changeRows) => {
          console.log("全选/取消全选:", selected, selectedRows.length);
        },
        getCheckboxProps: (record) => ({
          disabled: (record.deviceCount || 0) > 0, // 有设备使用的标签不能批量删除
          name: record.tagName,
        }),
      }}
      pagination={{
        pageSize: DEVICE_TAG_TABLE_CONFIG.DEFAULT_PAGE_SIZE,
        showSizeChanger: true,
        showQuickJumper: true,
      }}
    />
  );
};

export default DeviceTagTable;
