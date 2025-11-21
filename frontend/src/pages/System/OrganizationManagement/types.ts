// 组织管理页面内部类型定义

// 组织选择状态
export interface OrganizationSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: API.OrganizationBrief[];
}

// 模态框状态
export interface ModalState {
  create: boolean;
  edit: boolean;
  members: boolean;
  move: boolean;
  bind: boolean;
}

// 页面视图类型
export type ViewType = "tree" | "table";

// 树节点数据
export interface TreeNodeData extends Omit<API.OrganizationTree, "children"> {
  key: string;
  title: string | React.ReactNode;
  children?: TreeNodeData[];
}

// 组织树节点选择事件
export interface TreeSelectInfo {
  selected: boolean;
  selectedNodes: TreeNodeData[];
  node: TreeNodeData;
  event: "select";
}

// 拖拽信息
export interface TreeDragInfo {
  event: React.MouseEvent;
  node: TreeNodeData;
}

// 拖拽放置信息
export interface TreeDropInfo {
  event: React.MouseEvent;
  node: TreeNodeData;
  dragNode: TreeNodeData;
  dragNodesKeys: React.Key[];
  dropPosition: number;
  dropToGap: boolean;
}

// 组织成员选择状态
export interface MemberSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: API.UserOrgRelation[];
}

// 组织成员模态框状态
export interface MemberModalState {
  add: boolean;
  edit: boolean;
}

// 移动组织表单数据
export interface MoveOrganizationFormData {
  newParentId: string;
  newSortOrder?: number;
}

// 添加成员表单数据
export interface AddMemberFormData {
  userId: string;
  relationType: string;
  positionTitle?: string;
  isPrimary: boolean;
}

// 编辑成员表单数据
export interface EditMemberFormData {
  relationType?: string;
  positionTitle?: string;
  isPrimary?: boolean;
  status?: string;
}
