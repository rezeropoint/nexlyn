/**
 * 逻辑图配置管理 - 常量定义
 */

/** 表格配置 */
export const TABLE_CONFIG = {
  DEFAULT_PAGE_SIZE: 10, // 默认分页大小
  SCROLL_X: 1400, // 表格横向滚动宽度
};

/** 节点类型选项 */
export const NODE_TYPE_OPTIONS = [
  { label: '入口节点 (Entry)', value: 'entry' },
  { label: '动作节点 (Action)', value: 'action' },
  { label: '过滤节点 (Filter)', value: 'filter' },
  { label: '状态机节点 (StateMachine)', value: 'state_machine' },
  { label: '条件节点 (Condition)', value: 'condition' },
];

/** 节点类型颜色映射 */
export const NODE_TYPE_COLORS: Record<string, string> = {
  entry: 'blue',
  action: 'green',
  filter: 'orange',
  state_machine: 'purple',
  condition: 'cyan',
};

/** 逻辑块类别 */
export const BLOCK_CATEGORIES: Record<string, string> = {
  standard: '标准逻辑块',
  skylark: 'Skylark集成',
  custom: '自定义逻辑块',
};

/** 默认版本号 */
export const DEFAULT_VERSION = 'v1';

/** 默认节点数据（不包含ID，由组件自动生成临时ID） */
export const DEFAULT_NODE = {
  id: '', // 临时ID由NodeListEditor自动生成（格式: client:node-xxx）
  type: 'entry',
  blockType: '',
  blockVersion: '',
  isEntryPoint: false,
  subscribedInfoAtomTypeIDs: [],
  subscribedSource: '',
  subscribedLabels: [],
  blockConfig: '{}',
};

/** 默认边数据（不包含ID，由组件自动生成临时ID） */
export const DEFAULT_EDGE = {
  id: '', // 临时ID由EdgeListEditor自动生成（格式: client:edge-xxx）
  sourceID: '',
  targetID: '',
  condition: '',
};
