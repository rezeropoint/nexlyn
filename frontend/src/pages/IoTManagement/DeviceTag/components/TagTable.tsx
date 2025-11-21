import { type ListTagsRequest, listTags } from "@/services/iot";
import { DeleteOutlined, PlusOutlined } from "@ant-design/icons";
import {
  type ActionType,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { Badge, Button, Modal, Tag, message, theme } from "antd";
import React from "react";
import { TABLE_CONFIG } from "../constants";
import type { DeviceTagSummary, SelectionState } from "../types";
import styles from "./TagTable.less";
import TagTableActions from "./TagTableActions";

interface TagTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: SelectionState;
  onSelectionChange: (state: Partial<SelectionState>) => void;
  onCreateTag: () => void;
  onEditTag: (record: DeviceTagSummary) => void;
  onDeleteTag: (record: DeviceTagSummary) => void;
  onBatchDelete: (tags: DeviceTagSummary[]) => void;
}

/**
 * 标签表格组件
 */
const TagTable: React.FC<TagTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateTag,
  onEditTag,
  onDeleteTag,
  onBatchDelete,
}) => {
  const { token } = theme.useToken();
  const { selectedRowKeys, selectedRows } = selectionState;

  // 定义表格列
  const columns: ProColumns<DeviceTagSummary>[] = [
    {
      title: "标签名称",
      dataIndex: "name",
      width: 200,
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
      title: "标签颜色",
      dataIndex: "color",
      width: 150,
      hideInSearch: true,
      render: (_, record) => (
        <div className={styles.tagDisplay}>
          <Badge color={record.color || token.colorPrimary} />
          <Tag color={record.color || token.colorPrimary}>{record.name}</Tag>
        </div>
      ),
    },
    {
      title: "描述",
      dataIndex: "description",
      ellipsis: true,
      hideInSearch: true,
      render: (text) => text || "-",
    },
    {
      title: "操作",
      valueType: "option",
      width: 150,
      fixed: "right",
      render: (_, record) => (
        <TagTableActions
          record={record}
          onEdit={onEditTag}
          onDelete={onDeleteTag}
        />
      ),
    },
  ];

  // 工具栏渲染
  const toolBarRender = () => {
    const hasSelected = selectedRowKeys.length > 0;

    return [
      // 批量删除按钮
      hasSelected && (
        <Button
          key="batchDelete"
          danger
          icon={<DeleteOutlined />}
          onClick={() => {
            Modal.confirm({
              title: `确定要删除选中的 ${selectedRows.length} 个标签吗？`,
              content: "删除后将无法恢复",
              okText: "确定",
              cancelText: "取消",
              onOk: () => onBatchDelete(selectedRows),
            });
          }}
        >
          批量删除
        </Button>
      ),
      // 新建标签按钮
      <Button
        key="create"
        type="primary"
        icon={<PlusOutlined />}
        onClick={onCreateTag}
      >
        新建标签
      </Button>,
    ].filter(Boolean);
  };

  return (
    <ProTable<DeviceTagSummary, ListTagsRequest>
      headerTitle="标签列表"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: 120,
      }}
      scroll={{ x: TABLE_CONFIG.SCROLL_X }}
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
        try {
          const response = await listTags({
            page: params.current,
            pageSize: params.pageSize,
            name: params.name,
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
              response.msg || response.message || "获取标签列表失败"
            );
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        } catch (error) {
          console.error("获取标签列表失败:", error);
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

export default TagTable;
