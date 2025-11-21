import { MESSAGE } from "@/pages/EventManagement/constants";
import type {
  CreateOrgMappingRequest,
  OrgMapping,
} from "@/pages/EventManagement/types";
import {
  createOrgMapping,
  deleteOrgMapping,
  updateOrgMapping,
} from "@/services/eventhandler";
import { message } from "antd";
import { useState } from "react";

type ModalState = {
  create: boolean;
  edit: boolean;
};

export const useOrgMapping = () => {
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
  });
  const [currentRow, setCurrentRow] = useState<OrgMapping | undefined>(
    undefined
  );

  const setModalOpen = (type: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [type]: open }));
    // 只在真正关闭弹窗时清空数据（所有弹窗都关闭）
    if (!open && type === "edit") {
      setCurrentRow(undefined);
    }
    if (!open && type === "create") {
      setCurrentRow(undefined);
    }
  };

  const handleEditClick = (record: OrgMapping) => {
    setCurrentRow(record);
    setModalOpen("edit", true);
  };

  const handleCreate = async (values: CreateOrgMappingRequest) => {
    try {
      const response = await createOrgMapping(values);
      if (response.code === 0) {
        message.success(MESSAGE.CREATE_SUCCESS);
        setModalOpen("create", false);
        return true;
      }
      message.error(response.msg || MESSAGE.CREATE_FAILED);
      return false;
    } catch (_error) {
      message.error(MESSAGE.CREATE_FAILED);
      return false;
    }
  };

  const handleUpdate = async (values: any) => {
    // 从表单值或 currentRow 中获取 ID
    const id = values.id || currentRow?.id;
    if (!id) {
      message.error("缺少映射ID");
      return false;
    }

    try {
      const response = await updateOrgMapping(id, values);
      if (response.code === 0) {
        message.success(MESSAGE.UPDATE_SUCCESS);
        setModalOpen("edit", false);
        return true;
      }
      message.error(response.msg || MESSAGE.UPDATE_FAILED);
      return false;
    } catch (_error) {
      message.error(MESSAGE.UPDATE_FAILED);
      return false;
    }
  };

  const handleDelete = async (record: OrgMapping) => {
    try {
      const response = await deleteOrgMapping(record.id);
      if (response.code === 0) {
        message.success(MESSAGE.DELETE_SUCCESS);
        return true;
      }
      message.error(response.msg || MESSAGE.DELETE_FAILED);
      return false;
    } catch (_error) {
      message.error(MESSAGE.DELETE_FAILED);
      return false;
    }
  };

  return {
    currentRow,
    modalState,
    setModalOpen,
    handleCreate,
    handleUpdate,
    handleDelete,
    handleEditClick,
  };
};
