import {
  addOrganizationMember,
  removeOrganizationMember,
  updateOrganizationMember,
} from "@/services/organization";
import { message } from "antd";
import { useCallback, useState } from "react";
import type { MemberModalState, MemberSelectionState } from "../types";

/**
 * 组织成员管理业务逻辑Hook
 */
export const useOrganizationMembers = (orgId: string) => {
  // 当前编辑的成员记录
  const [currentMember, setCurrentMember] = useState<
    API.UserOrgRelation | undefined
  >();

  // 成员模态框状态
  const [memberModalState, setMemberModalState] = useState<MemberModalState>({
    add: false,
    edit: false,
  });

  // 成员选择状态
  const [memberSelectionState, setMemberSelectionState] =
    useState<MemberSelectionState>({
      selectedRowKeys: [],
      selectedRows: [],
    });

  // 设置成员模态框状态
  const setMemberModalOpen = useCallback(
    (type: keyof MemberModalState, open: boolean) => {
      setMemberModalState((prev) => ({ ...prev, [type]: open }));
    },
    []
  );

  // 处理编辑成员点击
  const handleEditMemberClick = useCallback(
    (record: API.UserOrgRelation) => {
      setCurrentMember(record);
      setMemberModalOpen("edit", true);
    },
    [setMemberModalOpen]
  );

  // 添加组织成员
  const handleAddMember = useCallback(
    async (values: {
      userId: string;
      relationType?: string;
      positionTitle?: string;
      isPrimary?: boolean;
    }): Promise<boolean> => {
      if (!orgId) return false;

      try {
        const response = await addOrganizationMember({ id: orgId }, values);
        if (response.code === 0) {
          message.success("成员添加成功");
          setMemberModalOpen("add", false);
          return true;
        } else {
          message.error(response.msg || "添加成员失败");
          return false;
        }
      } catch (error) {
        console.error("添加成员失败:", error);
        message.error("添加成员失败");
        return false;
      }
    },
    [orgId, setMemberModalOpen]
  );

  // 移除组织成员
  const handleRemoveMember = useCallback(
    async (member: API.UserOrgRelation) => {
      if (!orgId) return false;

      try {
        const response = await removeOrganizationMember({
          orgId,
          userId: member.userId,
        });
        if (response.code === 0) {
          message.success("成员移除成功");
          return true;
        } else {
          message.error(response.msg || "移除成员失败");
          return false;
        }
      } catch (error) {
        console.error("移除成员失败:", error);
        message.error("移除成员失败");
        return false;
      }
    },
    [orgId]
  );

  // 更新成员关系
  const handleUpdateMember = useCallback(
    async (values: {
      relationType?: string;
      positionTitle?: string;
      isPrimary?: boolean;
      status?: string;
    }): Promise<boolean> => {
      if (!orgId || !currentMember) return false;

      try {
        const response = await updateOrganizationMember(
          { orgId, userId: currentMember.userId },
          values
        );
        if (response.code === 0) {
          message.success("成员信息更新成功");
          setMemberModalOpen("edit", false);
          return true;
        } else {
          message.error(response.msg || "更新成员信息失败");
          return false;
        }
      } catch (error) {
        console.error("更新成员失败:", error);
        message.error("更新成员信息失败");
        return false;
      }
    },
    [orgId, currentMember, setMemberModalOpen]
  );

  // 批量移除成员
  const handleBatchRemoveMembers = useCallback(
    async (members: API.UserOrgRelation[]) => {
      if (!orgId) return false;

      try {
        const promises = members.map((member) =>
          removeOrganizationMember({
            orgId,
            userId: member.userId,
          })
        );
        const results = await Promise.all(promises);
        const successCount = results.filter(
          (result) => result.code === 0
        ).length;

        if (successCount === members.length) {
          message.success(`成功移除 ${successCount} 个成员`);
        } else if (successCount > 0) {
          message.warning(
            `成功移除 ${successCount} 个成员，${
              members.length - successCount
            } 个失败`
          );
        } else {
          message.error("批量移除失败");
        }

        // 清空选择状态
        setMemberSelectionState({ selectedRowKeys: [], selectedRows: [] });
        return successCount > 0;
      } catch (error) {
        console.error("批量移除成员失败:", error);
        message.error("批量移除失败");
        return false;
      }
    },
    [orgId]
  );

  // 批量更新成员关系类型
  const handleBatchUpdateMemberType = useCallback(
    async (members: API.UserOrgRelation[], relationType: string) => {
      if (!orgId) return false;

      try {
        const promises = members.map((member) =>
          updateOrganizationMember(
            { orgId, userId: member.userId },
            { relationType }
          )
        );
        const results = await Promise.all(promises);
        const successCount = results.filter(
          (result) => result.code === 0
        ).length;

        if (successCount === members.length) {
          message.success(`成功更新 ${successCount} 个成员的关系类型`);
        } else if (successCount > 0) {
          message.warning(
            `成功更新 ${successCount} 个成员，${
              members.length - successCount
            } 个失败`
          );
        } else {
          message.error("批量更新失败");
        }

        // 清空选择状态
        setMemberSelectionState({ selectedRowKeys: [], selectedRows: [] });
        return successCount > 0;
      } catch (error) {
        console.error("批量更新成员关系失败:", error);
        message.error("批量更新失败");
        return false;
      }
    },
    [orgId]
  );

  // 设置成员为负责人（一个组织只能有一个负责人）
  const handleSetAsManager = useCallback(
    async (member: API.UserOrgRelation) => {
      if (!orgId) return false;

      try {
        const response = await updateOrganizationMember(
          { orgId, userId: member.userId },
          { relationType: "manager", isPrimary: true }
        );
        if (response.code === 0) {
          message.success("已设置为组织负责人");
          return true;
        } else {
          message.error(response.msg || "设置负责人失败");
          return false;
        }
      } catch (error) {
        console.error("设置负责人失败:", error);
        message.error("设置负责人失败");
        return false;
      }
    },
    [orgId]
  );

  return {
    // 状态
    currentMember,
    memberModalState,
    memberSelectionState,

    // 操作方法
    setCurrentMember,
    setMemberModalOpen,
    setMemberSelectionState,

    // 业务操作
    handleAddMember,
    handleRemoveMember,
    handleUpdateMember,
    handleBatchRemoveMembers,
    handleBatchUpdateMemberType,
    handleEditMemberClick,
    handleSetAsManager,
  };
};
