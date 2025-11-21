import {
  getOrganizationTree,
  getUserOrganizations,
} from "@/services/organization";
import type { MessageInstance } from "antd/es/message/interface";
import { useCallback, useEffect, useState } from "react";

// 组织树节点接口（用于TreeSelect）
export interface OrgTreeNode {
  id: string;
  title: string;
  value: string;
  key: string;
  children?: OrgTreeNode[];
}

/**
 * 组织树管理hook
 * @param currentUser 当前用户信息
 * @param message message 实例
 */
export const useOrganization = (
  currentUser: API.CurrentUser | undefined,
  message: MessageInstance
) => {
  const [selectedOrg, setSelectedOrg] = useState<string>("");
  const [orgTreeData, setOrgTreeData] = useState<OrgTreeNode[]>([]);
  const [loadingOrg, setLoadingOrg] = useState(false);

  // 将API返回的组织树转换为TreeSelect需要的格式
  const convertToTreeData = useCallback(
    (organizations: API.OrganizationTree[]): OrgTreeNode[] => {
      return organizations.map((org) => ({
        id: org.id,
        key: org.id,
        value: org.id,
        title: org.name,
        children: org.children ? convertToTreeData(org.children) : [],
      }));
    },
    []
  );

  // 加载组织树数据
  const loadOrgTree = useCallback(async () => {
    setLoadingOrg(true);
    try {
      const response = await getOrganizationTree({});
      if (response.code === 0 && response.data) {
        const convertedData = convertToTreeData(response.data);
        setOrgTreeData(convertedData);
        return convertedData;
      } else {
        message.error(response.msg || "获取组织树失败");
        return [];
      }
    } catch (_error) {
      message.error("获取组织树失败");
      return [];
    } finally {
      setLoadingOrg(false);
    }
  }, [convertToTreeData]);

  // 获取用户的主组织
  const loadPrimaryOrg = useCallback(async () => {
    try {
      if (!currentUser || !currentUser.id) {
        return null;
      }

      const response = await getUserOrganizations({
        id: currentUser.id,
        isPrimary: true,
      });

      if (
        response.code === 0 &&
        response.data?.list &&
        response.data.list.length > 0
      ) {
        const primaryOrg = response.data.list[0];
        return primaryOrg.orgId;
      } else {
        return null;
      }
    } catch (_error) {
      return null;
    }
  }, [currentUser]);

  // 递归查找组织ID是否在树中存在
  const findOrgInTree = useCallback(
    (tree: OrgTreeNode[], orgId: string): boolean => {
      for (const node of tree) {
        if (node.id === orgId) {
          return true;
        }
        if (node.children && node.children.length > 0) {
          if (findOrgInTree(node.children, orgId)) {
            return true;
          }
        }
      }
      return false;
    },
    []
  );

  // 初始化组织数据
  useEffect(() => {
    const initOrgData = async () => {
      // 加载组织树
      const treeData = await loadOrgTree();

      // 获取用户主组织
      const primaryOrgId = await loadPrimaryOrg();

      // 设置选中的组织（验证组织是否存在）
      if (primaryOrgId && findOrgInTree(treeData, primaryOrgId)) {
        // 主组织ID有效，使用它
        setSelectedOrg(primaryOrgId);
      } else {
        // 主组织ID无效或不存在，使用组织树的第一个节点
        if (treeData.length > 0) {
          setSelectedOrg(treeData[0].id);
          if (primaryOrgId) {
            message.warning("您的主组织不存在，已切换到默认组织");
          }
        }
      }
    };

    initOrgData();
  }, [loadOrgTree, loadPrimaryOrg, findOrgInTree]);

  return {
    selectedOrg,
    setSelectedOrg,
    orgTreeData,
    loadingOrg,
  };
};
