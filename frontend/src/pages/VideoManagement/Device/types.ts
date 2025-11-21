import type { DeviceTag, NexlynDevice } from "@/services/video";

// 设备管理模态框状态
export interface DeviceModalState {
  bind: boolean;
  editAlias: boolean;
  editTags: boolean;
  edit: boolean; // 统一编辑表单
}

// 设备选择状态
export interface DeviceSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: NexlynDevice[];
}

// 设备绑定表单值
export interface DeviceBindFormValues {
  deviceId: string;
  deviceAlias: string;
  tags?: string[];
}

// 设备别名编辑表单值
export interface DeviceAliasFormValues {
  deviceAlias: string;
}

// 设备标签编辑表单值
export interface DeviceTagsFormValues {
  tags: string[];
}

// 统一的设备编辑表单值
export interface DeviceEditFormValues {
  deviceAlias: string;
  tags?: string[];
}

// 设备权限Hook返回类型
export interface UseDevicePermissionsReturn {
  canRead: boolean;
  canBind: boolean;
  canUnbind: boolean;
  canEdit: boolean;
  canControl: boolean;
  canViewUnbound: boolean;
}

// 设备管理Hook返回类型
export interface UseDeviceManagementReturn {
  // 状态
  currentDevice?: NexlynDevice;
  modalState: DeviceModalState;
  selectionState: DeviceSelectionState;
  availableTags: DeviceTag[];
  unboundDevices: NexlynDevice[];
  loadingUnbound: boolean;

  // 操作方法
  setCurrentDevice: (device?: NexlynDevice) => void;
  setModalOpen: (modal: keyof DeviceModalState, open: boolean) => void;
  setSelectionState: (state: Partial<DeviceSelectionState>) => void;
  refreshUnboundDevices: () => Promise<void>;
  refreshAvailableTags: () => Promise<void>;

  // 业务操作
  handleBind: (values: DeviceBindFormValues) => Promise<boolean>;
  handleUnbind: (device: NexlynDevice) => Promise<void>;
  handleUpdateAlias: (values: DeviceAliasFormValues) => Promise<boolean>;
  handleUpdateTags: (values: DeviceTagsFormValues) => Promise<boolean>;
  handleEditDevice: (values: DeviceEditFormValues) => Promise<boolean>;
  handleEditClick: (device: NexlynDevice) => void;
  handleEditAliasClick: (device: NexlynDevice) => void;
  handleEditTagsClick: (device: NexlynDevice) => void;
}

// 设备操作按钮配置
export interface DeviceActionConfig {
  showBind?: boolean;
  showUnbind?: boolean;
  showEdit?: boolean; // 统一编辑（合并别名和标签）
  showEditAlias?: boolean;
  showEditTags?: boolean;
  showSync?: boolean;
  showDelete?: boolean;
  showChannels?: boolean;
}

// 设备绑定状态筛选类型
export type DeviceBindingFilter = "all" | "bound" | "unbound";

// 设备过滤参数
export interface DeviceFilterParams {
  status?: "ON" | "OFF";
  keyword?: string;
  tags?: string[];
  bindingStatus?: DeviceBindingFilter;
}

// 设备汇总统计（简化版本）
export interface DeviceSummaryStats {
  deviceTotal: number;
  deviceOnline: number;
  deviceOffline: number;
}
