/**
 * 逻辑图配置表格组件
 */

import {
  type GetGraphConfigListRequest,
  type GraphConfigMetadata,
  listGraphConfigs,
} from '@/services/lynxmanager';
import { useModel } from '@@/exports';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { type ActionType, type ProColumns, ProTable } from '@ant-design/pro-components';
import { App, Button, Modal, Space, Switch, Tag, theme } from 'antd';
import dayjs from 'dayjs';
import React from 'react';
import { useNavigate } from '@@/exports';
import { TABLE_CONFIG } from '../constants';
import type { SelectionState } from '../types';
import GraphConfigTableActions from './GraphConfigTableActions';
import { getIconComponent, getIconColor } from '../iconConfig';

interface GraphConfigTableProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  selectionState: SelectionState;
  onSelectionChange: (state: Partial<SelectionState>) => void;
  onCreateConfig: () => void;
  onEditConfig: (record: GraphConfigMetadata) => void;
  onDeleteConfig: (record: GraphConfigMetadata) => void;
  onBatchDelete: (configs: GraphConfigMetadata[]) => void;
  onViewConfig?: (record: GraphConfigMetadata) => void;
  onCopyConfig?: (record: GraphConfigMetadata) => void;
}

/**
 * 逻辑图配置表格组件
 */
const GraphConfigTable: React.FC<GraphConfigTableProps> = ({
  actionRef,
  selectionState,
  onSelectionChange,
  onCreateConfig,
  onEditConfig: _onEditConfig,
  onDeleteConfig,
  onBatchDelete,
  onViewConfig,
  onCopyConfig,
}) => {
  const { token } = theme.useToken();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const { selectedRowKeys, selectedRows } = selectionState;
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;

  // 定义表格列
  const columns: ProColumns<GraphConfigMetadata>[] = [
    {
      title: '图名称',
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
      render: (text, record) => (
        <a
          onClick={(e) => {
            e.stopPropagation();
            if (onViewConfig) {
              onViewConfig(record);
            } else {
              navigate(`/logic-engine/graph-config/${record.id}`);
            }
          }}
        >
          {text}
        </a>
      ),
    },
    {
      title: '图标',
      dataIndex: 'icon',
      width: 80,
      hideInSearch: true,
      align: 'center',
      render: (_, record) => {
        const IconComponent = record.icon ? getIconComponent(record.icon) : null;
        const iconColor = record.iconColor || getIconColor(record.icon, token);
        return IconComponent ? (
          <IconComponent size={20} color={iconColor} />
        ) : (
          <span style={{ color: 'var(--ant-color-text-tertiary)' }}>-</span>
        );
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
      title: '描述',
      dataIndex: 'description',
      width: 250,
      ellipsis: true,
      hideInSearch: true,
      render: (text) => text || <span style={{ color: 'var(--ant-color-text-tertiary)' }}>-</span>,
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
      title: '启用状态',
      dataIndex: 'isEnabled',
      width: 100,
      render: (_, record) => (
        <Switch
          checked={record.isEnabled}
          disabled
          checkedChildren="已启用"
          unCheckedChildren="未启用"
        />
      ),
      valueType: 'select',
      valueEnum: {
        true: { text: '已启用', status: 'Success' },
        false: { text: '未启用', status: 'Default' },
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      hideInSearch: true,
      render: (text: any) => (text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '操作',
      valueType: 'option',
      width: 240,
      fixed: 'right',
      render: (_, record) => (
        <GraphConfigTableActions
          record={record}
          onView={
            onViewConfig ||
            ((rec) => navigate(`/logic-engine/graph-config/${rec.id}`))
          }
          onCopy={onCopyConfig || (() => {})}
          onDelete={onDeleteConfig}
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
              title: `确定要删除选中的 ${selectedRows.length} 个逻辑图配置吗？`,
              content: '删除后将无法恢复',
              okText: '确定',
              cancelText: '取消',
              okButtonProps: { danger: true },
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
        onClick={onCreateConfig}
      >
        新建逻辑图配置
      </Button>,
    ].filter(Boolean);
  };

  return (
    <ProTable<GraphConfigMetadata, GetGraphConfigListRequest>
      headerTitle="逻辑图配置列表"
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

          const response = await listGraphConfigs({
            page: params.current,
            pageSize: params.pageSize,
            tenantId,
            name: params.name,
            tags: params.tags,
            isEnabled: params.isEnabled,
          });

          if (response.code === 0 && response.data) {
            return {
              data: response.data.list || [],
              success: true,
              total: response.total || 0,
            };
          } else {
            message.error(response.msg || response.message || '获取逻辑图配置列表失败');
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        } catch (error) {
          console.error('获取逻辑图配置列表失败:', error);
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

export default GraphConfigTable;
