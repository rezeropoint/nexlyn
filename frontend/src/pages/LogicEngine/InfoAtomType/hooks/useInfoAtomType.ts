/**
 * 信息原子类型管理业务逻辑Hook
 */

import {
  createInfoAtomType,
  deleteInfoAtomType,
  type InfoAtomType,
  updateInfoAtomType,
} from '@/services/lynxmanager';
import { App } from 'antd';
import { useState } from 'react';
import type { ModalState, SelectionState } from '../types';

export function useInfoAtomType() {
  const { message, modal } = App.useApp();
  // 当前编辑的信息原子类型
  const [currentRow, setCurrentRow] = useState<InfoAtomType | undefined>();

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
   * 创建信息原子类型
   */
  const handleCreate = async (values: any): Promise<boolean> => {
    try {
      const response = await createInfoAtomType(values);

      if (response.code === 0) {
        message.success('信息原子类型创建成功');
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
   * 更新信息原子类型
   */
  const handleUpdate = async (values: any): Promise<boolean> => {
    if (!currentRow) {
      message.error('未选择信息原子类型');
      return false;
    }

    try {
      const response = await updateInfoAtomType(currentRow.id, values);

      if (response.code === 0) {
        message.success('信息原子类型更新成功');
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
   * 删除信息原子类型
   */
  const handleDelete = async (record: InfoAtomType) => {
    modal.confirm({
      title: '确认删除',
      content: `确定要删除信息原子类型"${record.name}"吗？`,
      okText: '确定',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const response = await deleteInfoAtomType(record.id);

          if (response.code === 0) {
            message.success('信息原子类型删除成功');
          } else {
            message.error(response.msg || response.message || '删除失败');
          }
        } catch (error: any) {
          message.error(`删除失败: ${error.message || '未知错误'}`);
        }
      },
    });
  };

  /**
   * 批量删除信息原子类型
   */
  const handleBatchDelete = async (types: InfoAtomType[]) => {
    modal.confirm({
      title: '确认批量删除',
      content: `确定要删除选中的 ${types.length} 个信息原子类型吗？`,
      okText: '确定',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          let successCount = 0;
          let failCount = 0;

          for (const type of types) {
            try {
              const response = await deleteInfoAtomType(type.id);
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
            message.success(`成功删除 ${successCount} 个信息原子类型`);
          }
          if (failCount > 0) {
            message.error(`${failCount} 个信息原子类型删除失败`);
          }

          // 清空选择
          setSelectionState({
            selectedRowKeys: [],
            selectedRows: [],
          });
        } catch (error: any) {
          message.error(`批量删除失败: ${error.message || '未知错误'}`);
        }
      },
    });
  };

  /**
   * 处理编辑点击
   */
  const handleEditClick = (record: InfoAtomType) => {
    setCurrentRow(record);
    setModalOpen('edit', true);
  };

  return {
    // 状态
    currentRow,
    modalState,
    selectionState,

    // 设置方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleBatchDelete,
    handleEditClick,
  };
}
