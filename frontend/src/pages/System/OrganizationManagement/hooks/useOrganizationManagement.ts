import {
  createOrganization,
  deleteOrganization,
  moveOrganization,
  updateOrganization,
} from "@/services/organization";
import { useModel } from "@umijs/max";
import { message } from "antd";
import { useCallback, useState } from "react";
import type {
  ModalState,
  OrganizationSelectionState,
  ViewType,
} from "../types";

/**
 * 组织管理主要业务逻辑Hook
 */
export const useOrganizationManagement = () => {
  const { initialState } = useModel("@@initialState");
  const currentUser = initialState?.currentUser;

  // 状态管理
  const [currentRow, setCurrentRow] = useState<API.Organization | undefined>();
  const [viewType, setViewType] = useState<ViewType>("tree");
  const [selectedOrgId, setSelectedOrgId] = useState<string>("");

  // 模态框状态
  const [modalState, setModalState] = useState<ModalState>({
    create: false,
    edit: false,
    members: false,
    move: false,
    bind: false,
  });

  // 选择状态
  const [selectionState, setSelectionState] =
    useState<OrganizationSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });

  // 设置模态框开关状态
  const setModalOpen = useCallback((type: keyof ModalState, open: boolean) => {
    setModalState((prev) => ({ ...prev, [type]: open }));
  }, []);

  // 处理编辑点击
  const handleEditClick = useCallback(
    (record: API.Organization | API.OrganizationBrief) => {
      setCurrentRow(record as API.Organization);
      setModalOpen("edit", true);
    },
    [setModalOpen]
  );

  // 处理成员管理点击
  const handleMembersClick = useCallback(
    (record: API.Organization | API.OrganizationBrief) => {
      setCurrentRow(record as API.Organization);
      setSelectedOrgId(record.id);
      setModalOpen("members", true);
    },
    [setModalOpen]
  );

  // 处理移动点击
  const handleMoveClick = useCallback(
    (record: API.Organization | API.OrganizationBrief) => {
      setCurrentRow(record as API.Organization);
      setModalOpen("move", true);
    },
    [setModalOpen]
  );

  // 创建组织
  const handleCreate = useCallback(
    async (values: API.CreateOrganizationRequest): Promise<boolean> => {
      try {
        const response = await createOrganization(values);
        if (response.code === 0) {
          message.success("组织创建成功");
          setModalOpen("create", false);
          return true;
        } else {
          message.error(response.msg || "创建组织失败");
          return false;
        }
      } catch (error) {
        console.error("创建组织失败:", error);
        message.error("创建组织失败");
        return false;
      }
    },
    [setModalOpen]
  );

  // 更新组织
  const handleUpdate = useCallback(
    async (values: API.UpdateOrganizationRequest): Promise<boolean> => {
      if (!currentRow) return false;

      try {
        const response = await updateOrganization(
          { id: currentRow.id },
          values
        );
        if (response.code === 0) {
          message.success("组织信息更新成功");
          setModalOpen("edit", false);
          return true;
        } else {
          message.error(response.msg || "更新组织信息失败");
          return false;
        }
      } catch (error) {
        console.error("更新组织失败:", error);
        message.error("更新组织信息失败");
        return false;
      }
    },
    [currentRow, setModalOpen]
  );

  // 删除组织
  const handleDelete = useCallback(
    async (
      record: API.Organization | API.OrganizationBrief,
      forceDelete = false
    ) => {
      try {
        const response = await deleteOrganization(
          { id: record.id },
          { forceDelete }
        );
        if (response.code === 0) {
          const deletedCount = response.data?.deletedCount || 1;
          if (deletedCount > 1) {
            message.success(`成功删除 ${deletedCount} 个组织（包含子组织）`);
          } else {
            message.success("组织删除成功");
          }
          return true;
        } else {
          message.error(response.msg || "删除组织失败");
          return false;
        }
      } catch (error) {
        console.error("删除组织失败:", error);
        message.error("删除组织失败");
        return false;
      }
    },
    []
  );

  // 移动组织
  const handleMove = useCallback(
    async (values: {
      newParentId: string;
      newSortOrder?: number;
    }): Promise<boolean> => {
      if (!currentRow) return false;

      try {
        const response = await moveOrganization({ id: currentRow.id }, values);
        if (response.code === 0) {
          message.success("组织移动成功");
          setModalOpen("move", false);
          return true;
        } else {
          message.error(response.msg || "移动组织失败");
          return false;
        }
      } catch (error) {
        console.error("移动组织失败:", error);
        message.error("移动组织失败");
        return false;
      }
    },
    [currentRow, setModalOpen]
  );

  // 批量删除
  const handleBatchDelete = useCallback(
    async (organizations: API.OrganizationBrief[], forceDelete = false) => {
      try {
        const promises = organizations.map((org) =>
          deleteOrganization({ id: org.id }, { forceDelete })
        );
        const results = await Promise.all(promises);
        const successCount = results.filter(
          (result) => result.code === 0
        ).length;

        if (successCount === organizations.length) {
          message.success(`成功删除 ${successCount} 个组织`);
        } else if (successCount > 0) {
          message.warning(
            `成功删除 ${successCount} 个组织，${
              organizations.length - successCount
            } 个失败`
          );
        } else {
          message.error("批量删除失败");
        }

        // 清空选择状态
        setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        return successCount > 0;
      } catch (error) {
        console.error("批量删除组织失败:", error);
        message.error("批量删除失败");
        return false;
      }
    },
    []
  );

  // 切换组织状态
  const handleToggleStatus = useCallback(
    async (record: API.Organization | API.OrganizationBrief) => {
      const newStatus = record.status === "active" ? "inactive" : "active";
      try {
        const response = await updateOrganization(
          { id: record.id },
          { status: newStatus }
        );
        if (response.code === 0) {
          message.success(
            `组织${newStatus === "active" ? "启用" : "停用"}成功`
          );
          return true;
        } else {
          message.error(response.msg || "状态更新失败");
          return false;
        }
      } catch (error) {
        console.error("切换组织状态失败:", error);
        message.error("状态更新失败");
        return false;
      }
    },
    []
  );

  // 批量切换状态
  const handleBatchToggleStatus = useCallback(
    async (
      organizations: API.OrganizationBrief[],
      targetStatus: "active" | "inactive"
    ) => {
      try {
        const promises = organizations.map((org) =>
          updateOrganization({ id: org.id }, { status: targetStatus })
        );
        const results = await Promise.all(promises);
        const successCount = results.filter(
          (result) => result.code === 0
        ).length;

        if (successCount === organizations.length) {
          message.success(
            `成功${
              targetStatus === "active" ? "启用" : "停用"
            } ${successCount} 个组织`
          );
        } else if (successCount > 0) {
          message.warning(
            `成功操作 ${successCount} 个组织，${
              organizations.length - successCount
            } 个失败`
          );
        } else {
          message.error("批量操作失败");
        }

        // 清空选择状态
        setSelectionState({ selectedRowKeys: [], selectedRows: [] });
        return successCount > 0;
      } catch (error) {
        console.error("批量切换状态失败:", error);
        message.error("批量操作失败");
        return false;
      }
    },
    []
  );

  return {
    // 状态
    currentUser,
    currentRow,
    modalState,
    selectionState,
    viewType,
    selectedOrgId,

    // 操作方法
    setModalOpen,
    setSelectionState,
    setViewType,
    setSelectedOrgId,
    setCurrentRow,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleMove,
    handleBatchDelete,
    handleToggleStatus,
    handleBatchToggleStatus,
    handleEditClick,
    handleMembersClick,
    handleMoveClick,
  };
};
