import { getOrganizationTree } from "@/services/organization";
import { message } from "antd";
import { useCallback, useEffect, useState } from "react";
import type { TreeDropInfo, TreeNodeData, TreeSelectInfo } from "../types";

/**
 * 组织树相关业务逻辑Hook
 */
export const useOrganizationTree = () => {
  const [treeData, setTreeData] = useState<TreeNodeData[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [autoExpandParent, setAutoExpandParent] = useState<boolean>(true);
  const [loading, setLoading] = useState<boolean>(false);

  // 将组织树数据转换为Tree组件需要的格式
  const convertToTreeData = useCallback(
    (organizations: API.OrganizationTree[]): TreeNodeData[] => {
      return organizations.map((org) => ({
        ...org,
        key: org.id,
        title: `${org.name} (${org.code})`,
        children: org.children ? convertToTreeData(org.children) : [],
      }));
    },
    []
  );

  // 获取所有节点的key
  const getAllKeys = useCallback((nodes: TreeNodeData[]): React.Key[] => {
    const keys: React.Key[] = [];
    const traverse = (nodes: TreeNodeData[]) => {
      nodes.forEach((node) => {
        keys.push(node.key);
        if (node.children && node.children.length > 0) {
          traverse(node.children);
        }
      });
    };
    traverse(nodes);
    return keys;
  }, []);

  // 加载组织树数据
  const loadTreeData = useCallback(
    async (params?: { tenantId?: string; rootId?: string }) => {
      setLoading(true);
      try {
        const response = await getOrganizationTree(params || {});
        if (response.code === 0 && response.data) {
          const convertedData = convertToTreeData(response.data);
          setTreeData(convertedData);

          // 默认展开第一层节点
          if (convertedData.length > 0) {
            const firstLevelKeys = convertedData.map((item) => item.key);
            setExpandedKeys(firstLevelKeys);
          }
        } else {
          message.error(response.msg || "获取组织树失败");
        }
      } catch (error) {
        console.error("获取组织树失败:", error);
        message.error("获取组织树失败");
      } finally {
        setLoading(false);
      }
    },
    [convertToTreeData]
  );

  // 树节点选择事件处理
  const onSelect = useCallback(
    (selectedKeys: React.Key[], info: TreeSelectInfo) => {
      console.log("树节点选择:", selectedKeys, info);
      setSelectedKeys(selectedKeys);
    },
    []
  );

  // 树节点展开/收起事件处理
  const onExpand = useCallback((expandedKeys: React.Key[]) => {
    console.log("树节点展开:", expandedKeys);
    setExpandedKeys(expandedKeys);
    setAutoExpandParent(false);
  }, []);

  // 树节点拖拽开始事件
  const onDragStart = useCallback((info: any) => {
    console.log("拖拽开始:", info);
  }, []);

  // 树节点拖拽进入事件
  const onDragEnter = useCallback((info: any) => {
    console.log("拖拽进入:", info);
    setExpandedKeys(info.expandedKeys);
  }, []);

  // 树节点拖拽结束（放置）事件
  const onDrop = useCallback((info: TreeDropInfo) => {
    console.log("拖拽放置:", info);
    const { node, dragNode, dropToGap } = info;

    // TODO: 实现拖拽逻辑，调用移动组织API
    // 这里需要根据dropPosition和dropToGap判断是放到节点内部还是兄弟位置
    const dropKey = node.key;
    const dragKey = dragNode.key;

    console.log(
      `将组织 ${dragKey} 移动到 ${dropKey} ${dropToGap ? "旁边" : "内部"}`
    );

    // 暂时只是打印日志，实际实现需要调用移动API
    message.info("拖拽移动功能开发中...");
  }, []);

  // 展开所有节点
  const expandAll = useCallback(() => {
    const allKeys = getAllKeys(treeData);
    setExpandedKeys(allKeys);
    setAutoExpandParent(false);
  }, [treeData, getAllKeys]);

  // 收起所有节点
  const collapseAll = useCallback(() => {
    setExpandedKeys([]);
    setAutoExpandParent(false);
  }, []);

  // 搜索节点
  const searchNode = useCallback(
    (searchValue: string) => {
      if (!searchValue) {
        setExpandedKeys([]);
        setAutoExpandParent(true);
        return;
      }

      const findMatchedKeys = (
        nodes: TreeNodeData[],
        matchedKeys: React.Key[] = []
      ): React.Key[] => {
        nodes.forEach((node) => {
          if (
            node.name.toLowerCase().includes(searchValue.toLowerCase()) ||
            node.code.toLowerCase().includes(searchValue.toLowerCase())
          ) {
            matchedKeys.push(node.key);
          }
          if (node.children && node.children.length > 0) {
            findMatchedKeys(node.children, matchedKeys);
          }
        });
        return matchedKeys;
      };

      const matchedKeys = findMatchedKeys(treeData);
      setExpandedKeys(matchedKeys);
      setAutoExpandParent(true);

      if (matchedKeys.length === 0) {
        message.warning("没有找到匹配的组织");
      }
    },
    [treeData]
  );

  // 根据ID查找节点
  const findNodeById = useCallback(
    (id: string): TreeNodeData | null => {
      const findNode = (nodes: TreeNodeData[]): TreeNodeData | null => {
        for (const node of nodes) {
          if (node.id === id) {
            return node;
          }
          if (node.children && node.children.length > 0) {
            const found = findNode(node.children);
            if (found) return found;
          }
        }
        return null;
      };
      return findNode(treeData);
    },
    [treeData]
  );

  // 刷新树数据
  const refreshTree = useCallback(() => {
    loadTreeData();
  }, [loadTreeData]);

  // 初始化加载数据
  useEffect(() => {
    loadTreeData();
  }, [loadTreeData]);

  return {
    // 数据状态
    treeData,
    expandedKeys,
    selectedKeys,
    autoExpandParent,
    loading,

    // 事件处理
    onSelect,
    onExpand,
    onDragStart,
    onDragEnter,
    onDrop,

    // 操作方法
    loadTreeData,
    expandAll,
    collapseAll,
    searchNode,
    findNodeById,
    refreshTree,
  };
};
