/**
 * 逻辑图配置管理 - 页面类型定义
 */

import type { EdgeConfig, GraphConfigMetadata, NodeConfig } from '@/services/lynxmanager';

/** 模态框状态 */
export interface ModalState {
  create: boolean; // 创建模态框
  edit: boolean; // 编辑模态框
}

/** 选择状态 */
export interface SelectionState {
  selectedRowKeys: React.Key[]; // 选中的行键
  selectedRows: GraphConfigMetadata[]; // 选中的行数据
}

/** 节点编辑项（用于NodeListEditor） */
export interface NodeItem extends NodeConfig {
  key?: string; // React列表key（可选，自动生成）
}

/** 边编辑项（用于EdgeListEditor） */
export interface EdgeItem extends EdgeConfig {
  key?: string; // React列表key（可选，自动生成）
}
