/**
 * 逻辑图配置管理主页面
 *
 * 组件结构：
 * - 主页面负责组件编排和状态协调
 * - 业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */

import { AppstoreOutlined, UnorderedListOutlined } from '@ant-design/icons';
import { type ActionType, PageContainer } from '@ant-design/pro-components';
import { Segmented } from 'antd';
import React, { useRef } from 'react';
import CreateGraphConfigForm from './components/CreateGraphConfigForm';
import EditGraphConfigForm from './components/EditGraphConfigForm';
import GraphConfigCardView from './components/GraphConfigCardView';
import GraphConfigTable from './components/GraphConfigTable';
import { useGraphConfig } from './hooks/useGraphConfig';

const GraphConfigManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,
    selectionState,
    viewMode,

    // 设置方法
    setModalOpen,
    setSelectionState,
    setViewMode,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleBatchDelete,
    handleEditClick,
    handleView,
    handleCopy,
  } = useGraphConfig();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 操作处理（包含表格刷新）
  const handleOperation = {
    delete: async (record: any) => {
      const success = await handleDelete(record);
      if (success) {
        handleTableRefresh();
      }
    },
    batchDelete: async (configs: any[]) => {
      const success = await handleBatchDelete(configs);
      if (success) {
        handleTableRefresh();
      }
    },
    copy: async (record: any) => {
      const success = await handleCopy(record);
      if (success) {
        handleTableRefresh();
      }
    },
  };

  return (
    <PageContainer
      subTitle="配置和管理逻辑图，定义节点、边和执行流程"
      extra={
        <Segmented
          value={viewMode}
          onChange={setViewMode}
          options={[
            {
              label: '卡片视图',
              value: 'card',
              icon: <AppstoreOutlined />,
            },
            {
              label: '表格视图',
              value: 'table',
              icon: <UnorderedListOutlined />,
            },
          ]}
        />
      }
    >
      {/* 条件渲染：卡片视图或表格视图 */}
      {viewMode === 'card' ? (
        <GraphConfigCardView
          actionRef={actionRef}
          onView={handleView}
          onCopy={handleOperation.copy}
          onDelete={handleOperation.delete}
          onCreateConfig={() => setModalOpen('create', true)}
        />
      ) : (
        <GraphConfigTable
          actionRef={actionRef}
          selectionState={selectionState}
          onSelectionChange={(state) => setSelectionState((prev) => ({ ...prev, ...state }))}
          onCreateConfig={() => setModalOpen('create', true)}
          onEditConfig={handleEditClick}
          onViewConfig={handleView}
          onCopyConfig={handleOperation.copy}
          onDeleteConfig={handleOperation.delete}
          onBatchDelete={handleOperation.batchDelete}
        />
      )}

      {/* 创建表单 */}
      <CreateGraphConfigForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen('create', open)}
        onFinish={handleCreateSuccess}
      />

      {/* 编辑表单 */}
      <EditGraphConfigForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen('edit', open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />
    </PageContainer>
  );
};

export default GraphConfigManagement;
