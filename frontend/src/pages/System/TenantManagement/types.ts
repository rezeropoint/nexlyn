// 本地类型定义
import type {
  CreateTenantRequest,
  Tenant,
  UpdateTenantRequest,
  UpdateTenantStatusRequest,
} from "@/services/tenant";

// 选择行状态类型
export interface TenantSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: Tenant[];
}

// 模态框状态类型
export interface ModalState {
  create: boolean;
  edit: boolean;
  statusManage: boolean;
}

// 选项数据类型
export interface OptionsState {
  tags: { label: string; value: string }[];
}

// 租户操作类型
export type TenantOperation =
  | "create"
  | "edit"
  | "delete"
  | "updateStatus"
  | "batchDelete"
  | "batchToggleStatus";

// 批量操作参数类型
export interface BatchOperationParams {
  tenants: Tenant[];
  operation: "delete" | "activate" | "deactivate";
}

// 租户权限检查结果类型
export interface TenantPermissionCheck {
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
  canOperateTenant: (tenant: Tenant) => boolean;
}

// 表单提交参数类型
export interface FormSubmitParams {
  create: CreateTenantRequest;
  update: UpdateTenantRequest;
  updateStatus: UpdateTenantStatusRequest;
}

// Hook返回值类型定义
export interface UseTenantManagementReturn {
  // 状态
  currentUser: any;
  currentRow: Tenant | undefined;
  modalState: ModalState;
  selectionState: TenantSelectionState;

  // 操作方法
  setCurrentRow: (tenant: Tenant | undefined) => void;
  setModalOpen: (modal: keyof ModalState, open: boolean) => void;
  setSelectionState: (state: Partial<TenantSelectionState>) => void;

  // 业务操作
  handleCreate: (values: CreateTenantRequest) => Promise<boolean>;
  handleUpdate: (values: UpdateTenantRequest) => Promise<boolean>;
  handleDelete: (record: Tenant) => Promise<void>;
  handleUpdateStatus: (values: UpdateTenantStatusRequest) => Promise<boolean>;
  handleBatchDelete: (tenants: Tenant[]) => Promise<void>;
  handleBatchToggleStatus: (
    tenants: Tenant[],
    targetStatus: "active" | "inactive"
  ) => Promise<void>;
  handleToggleStatus: (record: Tenant) => Promise<void>;
  handleEditClick: (record: Tenant) => void;
  handleStatusClick: (record: Tenant) => void;
}

export interface UseTenantOptionsReturn {
  tagOptions: { label: string; value: string }[];
  loadTagOptions: () => Promise<void>;
}
