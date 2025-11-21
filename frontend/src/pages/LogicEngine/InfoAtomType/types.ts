/**
 * 信息原子类型管理 - 页面类型定义
 */

import type { InfoAtomType } from '@/services/lynxmanager';

/** 模态框状态 */
export interface ModalState {
  create: boolean; // 创建模态框
  edit: boolean; // 编辑模态框
}

/** 选择状态 */
export interface SelectionState {
  selectedRowKeys: React.Key[]; // 选中的行键
  selectedRows: InfoAtomType[]; // 选中的行数据
}
