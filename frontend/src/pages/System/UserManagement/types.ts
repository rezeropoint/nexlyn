// 本地类型定义

// 选择行状态类型
export interface UserSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: API.UserBrief[];
}

// 模态框状态类型
export interface ModalState {
  create: boolean;
  edit: boolean;
  resetPassword: boolean;
  bind: boolean;
}

// 选项数据类型
export interface OptionsState {
  tags: API.TagOption[];
  tenants: { label: string; value: string }[];
}

// 用户操作类型
export type UserOperation =
  | "create"
  | "edit"
  | "delete"
  | "resetPassword"
  | "updateStatus"
  | "batchDelete"
  | "batchToggleStatus";

// 批量操作参数类型
export interface BatchOperationParams {
  users: API.UserBrief[];
  operation: "delete" | "activate" | "deactivate";
}

// 用户权限检查结果类型
export interface UserPermissionCheck {
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
  canOperateUser: (user: API.UserBrief) => boolean;
}

// 表单提交参数类型
export interface FormSubmitParams {
  create: API.CreateUserRequest;
  update: API.UpdateUserRequest;
  resetPassword: { password: string };
  updateStatus: API.UpdateUserStatusRequest;
}

// Hook返回值类型定义
export interface UseUserManagementReturn {
  // 状态
  currentUser: API.User | undefined;
  currentRow: API.User | undefined;
  modalState: ModalState;
  selectionState: UserSelectionState;

  // 操作方法
  setCurrentRow: (user: API.User | undefined) => void;
  setModalOpen: (modal: keyof ModalState, open: boolean) => void;
  setSelectionState: (state: Partial<UserSelectionState>) => void;

  // 业务操作
  handleCreate: (values: API.CreateUserRequest) => Promise<boolean>;
  handleUpdate: (values: API.UpdateUserRequest) => Promise<boolean>;
  handleDelete: (record: API.UserBrief) => Promise<void>;
  handleResetPassword: (values: { password: string }) => Promise<boolean>;
  handleUpdateStatus: (values: API.UpdateUserStatusRequest) => Promise<boolean>;
  handleBatchDelete: (users: API.UserBrief[]) => Promise<void>;
  handleBatchToggleStatus: (
    users: API.UserBrief[],
    targetStatus: "active" | "inactive"
  ) => Promise<void>;
  handleToggleStatus: (record: API.UserBrief) => Promise<void>;
  handleEditClick: (record: API.UserBrief) => Promise<void>;
}

export interface UseUserOptionsReturn {
  tagOptions: API.TagOption[];
  tenantOptions: { label: string; value: string }[];
  loadTagOptions: () => Promise<void>;
  loadTenantOptions: () => Promise<void>;
}
