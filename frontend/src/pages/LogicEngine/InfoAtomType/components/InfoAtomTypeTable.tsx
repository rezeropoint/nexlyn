/**
 * 信息原子类型表格组件
 */

import {
  type GetInfoAtomTypeListRequest,
  type InfoAtomType,
  listInfoAtomTypes,
} from '@/services/lynxmanager';
import { useModel } from '@@/exports';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { type ActionType, type ProColumns, ProTable } from '@ant-design/pro-components';
import { App, Button, Modal, Space, Tag } from 'antd';
import dayjs from 'dayjs';
import React from 'react';
import { FIELD_TYPE_COLORS, TABLE_CONFIG } from '../constants';
import type { SelectionState } from '../types';
import InfoAtomTypeTableActions from './InfoAtomTypeTableActions';

interface InfoAtomTypeTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: SelectionState;
  onSelectionChange: (state: Partial<SelectionState>) => void;
  onCreateType: () => void;
  onEditType: (record: InfoAtomType) => void;
  onDeleteType: (record: InfoAtomType) => void;
  onBatchDelete: (types: InfoAtomType[]) => void;
}

/**
 * 信息原子类型表格组件
 */
const InfoAtomTypeTable: React.FC<InfoAtomTypeTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateType,
  onEditType,
  onDeleteType,
  onBatchDelete,
}) => {
  const { message } = App.useApp();
  const { selectedRowKeys, selectedRows } = selectionState;
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;

  // 定义表格列
  const columns: ProColumns<InfoAtomType>[] = [
    {
      title: '类型名称',
      dataIndex: 'name',
      width: 200,
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
      title: '版本',
      dataIndex: 'version',
      width: 100,
      hideInSearch: true,
      render: (text) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '标签',
      dataIndex: 'tags',
      width: 200,
      hideInSearch: true,
      render: (_, record) => (
        <Space wrap>
          {record.tags && record.tags.length > 0 ? (
            // Phase 2.5: 改为使用对象形式（tag.id, tag.name）
            record.tags.map((tag) => <Tag key={tag.id}>{tag.name}</Tag>)
          ) : (
            <span style={{ color: 'var(--ant-color-text-tertiary)' }}>-</span>
          )}
        </Space>
      ),
    },
    {
      title: '字段数量',
      dataIndex: 'fieldCount',
      width: 100,
      hideInSearch: true,
      render: (_, record) => (
        <Tag color="green">{record.dataFormat.fields.length} 个字段</Tag>
      ),
    },
    {
      title: '数据类型',
      dataIndex: 'dataPlural',
      width: 100,
      hideInSearch: true,
      render: (_, record) => (
        <Tag color={record.dataFormat.dataPlural ? 'orange' : 'default'}>
          {record.dataFormat.dataPlural ? '数组' : '单条'}
        </Tag>
      ),
    },
    {
      title: '字段类型',
      dataIndex: 'fieldTypes',
      width: 200,
      hideInSearch: true,
      render: (_, record) => (
        <Space wrap>
          {[...new Set(record.dataFormat.fields.map((f) => f.fieldType))].map(
            (fieldType) => (
              <Tag key={fieldType} color={FIELD_TYPE_COLORS[fieldType] || 'default'}>
                {fieldType}
              </Tag>
            ),
          )}
        </Space>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      hideInSearch: true,
      render: (text: any) =>
        text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
    {
      title: '操作',
      valueType: 'option',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <InfoAtomTypeTableActions
          record={record}
          onEdit={onEditType}
          onDelete={onDeleteType}
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
              title: `确定要删除选中的 ${selectedRows.length} 个信息原子类型吗？`,
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
        onClick={onCreateType}
      >
        新建信息原子类型
      </Button>,
    ].filter(Boolean);
  };

  return (
    <ProTable<InfoAtomType, GetInfoAtomTypeListRequest>
      headerTitle="信息原子类型列表"
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
          // 获取当前用户的租户ID（从tenantInfo对象中获取）
          const tenantId = currentUser?.tenantInfo?.tenantId;
          if (!tenantId) {
            // 租户信息未加载时返回空数据，不显示错误
            return {
              data: [],
              success: true,
              total: 0,
            };
          }

          const response = await listInfoAtomTypes({
            page: params.current,
            pageSize: params.pageSize,
            tenantId,
            name: params.name,
            tags: params.tags,
          });

          if (response.code === 0 && response.data) {
            return {
              data: response.data.list || [],
              success: true,
              total: response.total || 0,
            };
          } else {
            message.error(
              response.msg || response.message || '获取信息原子类型列表失败',
            );
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        } catch (error) {
          console.error('获取信息原子类型列表失败:', error);
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

export default InfoAtomTypeTable;
