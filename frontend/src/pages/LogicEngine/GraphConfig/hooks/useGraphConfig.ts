/**
 * 逻辑图配置管理业务逻辑Hook
 */

import {
  createGraphConfig,
  deleteGraphConfig,
  type GraphConfigMetadata,
  updateGraphConfig,
} from '@/services/lynxmanager';
import { useModel, useNavigate } from '@@/exports';
import { App } from 'antd';
import { useState } from 'react';
import type { ModalState, SelectionState } from '../types';

export function useGraphConfig() {
  const { message, modal } = App.useApp();
  const navigate = useNavigate();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;

  // 当前编辑的逻辑图配置
  const [currentRow, setCurrentRow] = useState<GraphConfigMetadata | undefined>();

  // 模态框状态
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
  });

  // 选择状态
  const [selectionState, setSelectionState] = useState<SelectionState>({
    selectedRowKeys: [],
    selectedRows: [],
  });

  // 视图模式状态 ('card' | 'table')
  const [viewMode, setViewMode] = useState<'card' | 'table'>('card');

  /**
   * 设置模态框打开状态
   */
  const setModalOpen = (type: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [type]: open }));
    if (!open) {
      setCurrentRow(undefined);
    }
  };

  /**
   * 创建逻辑图配置
   */
  const handleCreate = async (values: any): Promise<boolean> => {
    try {
      const response = await createGraphConfig(values);

      if (response.code === 0) {
        message.success('逻辑图配置创建成功');
        setModalOpen('create', false);
        return true;
      } else {
        message.error(response.msg || response.message || '创建失败');
        return false;
      }
    } catch (error: any) {
      message.error(`创建失败: ${error.message || '未知错误'}`);
      return false;
    }
  };

  /**
   * 更新逻辑图配置
   */
  const handleUpdate = async (values: any): Promise<boolean> => {
    if (!currentRow) {
      message.error('未选择逻辑图配置');
      return false;
    }

    try {
      const response = await updateGraphConfig(currentRow.id, values);

      if (response.code === 0) {
        message.success('逻辑图配置更新成功');
        setModalOpen('edit', false);
        return true;
      } else {
        message.error(response.msg || response.message || '更新失败');
        return false;
      }
    } catch (error: any) {
      message.error(`更新失败: ${error.message || '未知错误'}`);
      return false;
    }
  };

  /**
   * 删除逻辑图配置
   */
  const handleDelete = async (record: GraphConfigMetadata): Promise<boolean> => {
    return new Promise((resolve) => {
      modal.confirm({
        title: '确认删除',
        content: `确定要删除逻辑图配置"${record.name}"吗？`,
        okText: '确定',
        cancelText: '取消',
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            const response = await deleteGraphConfig(record.id);

            if (response.code === 0) {
              message.success('逻辑图配置删除成功');
              resolve(true);
            } else {
              message.error(response.msg || response.message || '删除失败');
              resolve(false);
            }
          } catch (error: any) {
            message.error(`删除失败: ${error.message || '未知错误'}`);
            resolve(false);
          }
        },
        onCancel: () => {
          resolve(false);
        },
      });
    });
  };

  /**
   * 批量删除逻辑图配置
   */
  const handleBatchDelete = async (configs: GraphConfigMetadata[]): Promise<boolean> => {
    return new Promise((resolve) => {
      modal.confirm({
        title: '确认批量删除',
        content: `确定要删除选中的 ${configs.length} 个逻辑图配置吗？`,
        okText: '确定',
        cancelText: '取消',
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            let successCount = 0;
            let failCount = 0;

            for (const config of configs) {
              try {
                const response = await deleteGraphConfig(config.id);
                if (response.code === 0) {
                  successCount++;
                } else {
                  failCount++;
                }
              } catch {
                failCount++;
              }
            }

            if (successCount > 0) {
              message.success(`成功删除 ${successCount} 个逻辑图配置`);
            }
            if (failCount > 0) {
              message.error(`${failCount} 个逻辑图配置删除失败`);
            }

            // 清空选择
            setSelectionState({
              selectedRowKeys: [],
              selectedRows: [],
            });

            resolve(successCount > 0);
          } catch (error: any) {
            message.error(`批量删除失败: ${error.message || '未知错误'}`);
            resolve(false);
          }
        },
        onCancel: () => {
          resolve(false);
        },
      });
    });
  };

  /**
   * 处理编辑点击
   */
  const handleEditClick = (record: GraphConfigMetadata) => {
    setCurrentRow(record);
    setModalOpen('edit', true);
  };

  /**
   * 处理查看详情（跳转到详情页）
   */
  const handleView = (record: GraphConfigMetadata) => {
    navigate(`/logic-engine/graph-config/${record.id}`);
  };

  /**
   * 处理复制逻辑图配置
   */
  const handleCopy = async (record: GraphConfigMetadata): Promise<boolean> => {
    const tenantId = currentUser?.tenantInfo?.tenantId;
    if (!tenantId) {
      message.error('无法获取租户信息');
      return false;
    }

    // 获取组织ID（organizationIds数组的第一个值是主组织ID）
    const orgId = currentUser?.organizationIds?.[0];
    if (!orgId) {
      message.error('无法获取用户组织信息');
      return false;
    }

    return new Promise((resolve) => {
      modal.confirm({
        title: '复制逻辑图配置',
        content: `确定要复制逻辑图配置"${record.name}"吗？新配置名称将添加"(副本)"后缀。`,
        okText: '确定',
        cancelText: '取消',
        onOk: async () => {
          try {
            // 创建副本数据
            const copyData = {
              tenantId,
              orgId, // 组织ID（用于权限控制）
              name: `${record.name}(副本)`,
              version: record.version,
              description: record.description,
              // Phase 2.5: 将 LynxTagSummary[] 转换为 string[] ID 数组
              tagIds: record.tags?.map(tag => tag.id),
              isEnabled: false, // 副本默认未启用
              nodes: [],
              edges: [],
            };

            const response = await createGraphConfig(copyData);

            if (response.code === 0) {
              message.success('逻辑图配置复制成功');
              resolve(true);
            } else {
              message.error(response.msg || response.message || '复制失败');
              resolve(false);
            }
          } catch (error: any) {
            message.error(`复制失败: ${error.message || '未知错误'}`);
            resolve(false);
          }
        },
        onCancel: () => {
          resolve(false);
        },
      });
    });
  };

  return {
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
  };
}
