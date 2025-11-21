/**
 * 标签管理 - 常量定义
 */

import { TagScope } from '@/services/lynxmanager';

/** 标签作用域选项 */
export const TAG_SCOPE_OPTIONS = {
  [TagScope.InfoAtom]: {
    label: '信息原子',
    value: TagScope.InfoAtom,
  },
  [TagScope.LogicGraph]: {
    label: '逻辑图',
    value: TagScope.LogicGraph,
  },
};

/** 标签作用域选项数组 */
export const TAG_SCOPE_ENUM = {
  info_atom: { text: '信息原子' },
  logic_graph: { text: '逻辑图' },
};

/** 分页配置 */
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = ['10', '20', '50', '100'];
