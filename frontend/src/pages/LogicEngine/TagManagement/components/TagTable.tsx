/**
 * 标签表格组件
 */

import {
  type ListTagsRequest,
  type LynxTag,
  listTags,
} from '@/services/lynxmanager';
import { useModel } from '@@/exports';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { type ActionType, type ProColumns, ProTable } from '@ant-design/pro-components';
import { App, Button, Modal, Tag } from 'antd';
import dayjs from 'dayjs';
import React from 'react';
import { DEFAULT_PAGE_SIZE, PAGE_SIZE_OPTIONS, TAG_SCOPE_ENUM } from '../constants';
import type { SelectionState } from '../types';
import TagTableActions from './TagTableActions';
import styles from './TagTable.less';

interface TagTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: SelectionState;
  onSelectionChange: (state: Partial<SelectionState>) => void;
  onCreateTag: () => void;
  onEditTag: (record: LynxTag) => void;
  onDeleteTag: (record: LynxTag) => void;
  onBatchDelete: (tags: LynxTag[]) => void;
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
  const { message } = App.useApp();
  const { selectedRowKeys, selectedRows } = selectionState;
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;

  // 定义表格列
  const columns: ProColumns<LynxTag>[] = [
    {
      title: '标签名称',
      dataIndex: 'name',
      width: 150,
      ellipsis: true,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
    },
    {
      title: '作用域',
      dataIndex: 'scope',
      width: 120,
      hideInTable: false,
      valueType: 'select',
      valueEnum: TAG_SCOPE_ENUM,
      render: (_, record) => {
        const scopeText = TAG_SCOPE_ENUM[record.scope as keyof typeof TAG_SCOPE_ENUM]?.text || record.scope;
        return <Tag color="blue">{scopeText}</Tag>;
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      width: 200,
      ellipsis: true,
      hideInSearch: true,
      render: (_, record) => (
        record.description || (
          <span style={{ color: 'var(--ant-color-text-tertiary)' }}>-</span>
        )
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      hideInSearch: true,
      render: (text: any) =>
        text ? dayjs.unix(text).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      width: 180,
      hideInSearch: true,
      render: (text: any) =>
        text ? dayjs.unix(text).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      valueType: 'option',
      width: 120,
      fixed: 'right',
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
              content: '删除后将无法恢复',
              okText: '确定',
              cancelText: '取消',
              onOk: () => onBatchDelete(selectedRows),
            });
          }}
        >
          批量删除
        </Button>
      ),
      // 新建按钮
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
    <ProTable<LynxTag, ListTagsRequest>
      className={styles.tagTable}
      headerTitle="标签列表"
      actionRef={actionRef}
      rowKey="id"
      search={{
        labelWidth: 120,
        defaultCollapsed: false,
      }}
      scroll={{ x: 1200 }}
      options={{
        reload: true,
        setting: true,
        density: true,
      }}
      pagination={{
        defaultPageSize: DEFAULT_PAGE_SIZE,
        pageSizeOptions: PAGE_SIZE_OPTIONS,
        showQuickJumper: true,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 条`,
      }}
      toolBarRender={toolBarRender}
      request={async (params) => {
        try {
          // 获取当前用户的租户ID
          const tenantId = currentUser?.tenantInfo?.tenantId;
          if (!tenantId) {
            return {
              data: [],
              success: true,
              total: 0,
            };
          }

          const response = await listTags({
            page: params.current,
            pageSize: params.pageSize,
            tenantId,
            scope: params.scope as any,
            keyword: (params as any).name,
          });

          if (response.code === 0 && response.data) {
            return {
              data: response.data.list || [],
              success: true,
              total: response.total || 0,
            };
          } else {
            message.error(
              response.msg || response.message || '获取标签列表失败',
            );
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        } catch (error) {
          console.error('获取标签列表失败:', error);
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
