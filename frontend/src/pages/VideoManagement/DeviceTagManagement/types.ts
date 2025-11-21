import type { DeviceTag } from "@/services/video";

// 设备标签管理模态框状态
export interface DeviceTagModalState {
  create: boolean;
  edit: boolean;
}

// 设备标签选择状态
export interface DeviceTagSelectionState {
  selectedRowKeys: React.Key[];
  selectedRows: DeviceTag[];
}

// 设备标签管理Hook返回类型
export interface UseDeviceTagManagementReturn {
  // 状态
  currentRow?: DeviceTag;
  modalState: DeviceTagModalState;
  selectionState: DeviceTagSelectionState;

  // 操作方法
  setCurrentRow: (row?: DeviceTag) => void;
  setModalOpen: (modal: keyof DeviceTagModalState, open: boolean) => void;
  setSelectionState: (state: Partial<DeviceTagSelectionState>) => void;

  // 业务操作
  handleCreate: (values: any) => Promise<boolean>;
  handleUpdate: (values: any) => Promise<boolean>;
  handleDelete: (record: DeviceTag) => Promise<void>;
  handleBatchDelete: (tags: DeviceTag[]) => Promise<void>;
  handleEditClick: (record: DeviceTag) => void;
}

// 设备标签权限Hook返回类型
export interface UseDeviceTagPermissionsReturn {
  canRead: boolean;
  canCreate: boolean;
  canUpdate: boolean;
  canDelete: boolean;
}
