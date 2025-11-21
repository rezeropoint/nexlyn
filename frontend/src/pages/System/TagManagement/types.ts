// 模态框状态
export interface ModalState {
  create: boolean;
  edit: boolean;
}

// 标签选择状态
export interface TagSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: API.TagDefinition[];
}

// Hook返回类型
export interface UseTagManagementReturn {
  // 状态
  currentRow?: API.TagDefinition;
  modalState: ModalState;
  selectionState: TagSelectionState;
  defaultScope: string;

  // 操作方法
  setCurrentRow: (row?: API.TagDefinition) => void;
  setModalOpen: (modal: keyof ModalState, open: boolean) => void;
  setSelectionState: (state: Partial<TagSelectionState>) => void;

  // 业务操作
  handleCreate: (values: API.CreateTagRequest) => Promise<boolean>;
  handleUpdate: (values: API.UpdateTagRequest) => Promise<boolean>;
  handleDelete: (record: API.TagDefinition) => Promise<void>;
  handleBatchDelete: (tags: API.TagDefinition[]) => Promise<void>;
  handleEditClick: (record: API.TagDefinition) => void;
}

// 权限Hook返回类型
export interface UseTagPermissionsReturn {
  // 综合标签权限（用户标签或租户标签）
  canRead: boolean;
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;

  // 用户标签权限（租户管理员及以上）
  canManageUserTags: boolean;
  canReadUserTags: boolean;
  canUpdateUserTags: boolean;
  canDeleteUserTags: boolean;

  // 租户标签权限（仅超级管理员）
  canManageTenantTags: boolean;
  canReadTenantTags: boolean;
  canUpdateTenantTags: boolean;
  canDeleteTenantTags: boolean;
}
