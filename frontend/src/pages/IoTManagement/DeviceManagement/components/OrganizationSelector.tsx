import { getOrganizationTree } from "@/services/organization";
import {
  CompressOutlined,
  ExpandAltOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { Button, Empty, message, Spin, Tree } from "antd";
import React, {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useState,
} from "react";
import styles from "./OrganizationSelector.less";

interface TreeNodeData {
  key: string;
  id: string;
  name: string;
  code: string;
  children?: TreeNodeData[];
  [key: string]: any;
}

interface OrganizationSelectorProps {
  selectedOrgId?: string;
  onSelect: (orgId: string) => void;
  hideToolbar?: boolean; // 是否隐藏工具栏（用于外部提供工具栏按钮的场景）
}

// 暴露给父组件的方法
export interface OrganizationSelectorRef {
  expandAll: () => void;
  collapseAll: () => void;
  refresh: () => void;
  getLoading: () => boolean;
}

/**
 * 组织树选择器组件
 * 用于设备管理页面，提供组织选择功能
 */
const OrganizationSelector = forwardRef<
  OrganizationSelectorRef,
  OrganizationSelectorProps
>(({ selectedOrgId, onSelect, hideToolbar = false }, ref) => {
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
        title: org.name, // 只显示组织名称，不显示组织代码
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
  const loadTreeData = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getOrganizationTree({});
      if (response.code === 0 && response.data) {
        const convertedData = convertToTreeData(response.data);
        setTreeData(convertedData);

        // 默认展开第一层节点
        if (convertedData.length > 0) {
          const firstLevelKeys = convertedData.map((item) => item.key);
          setExpandedKeys(firstLevelKeys);

          // 如果没有选中的组织，默认选中第一个根节点
          if (!selectedOrgId && convertedData.length > 0) {
            const firstOrgId = convertedData[0].id;
            setSelectedKeys([firstOrgId]);
            onSelect(firstOrgId);
          }
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
  }, [convertToTreeData, selectedOrgId, onSelect]);

  // 树节点选择事件处理
  const handleSelect = useCallback(
    (selectedKeys: React.Key[]) => {
      if (selectedKeys.length > 0) {
        const orgId = selectedKeys[0] as string;
        setSelectedKeys(selectedKeys);
        onSelect(orgId);
      }
    },
    [onSelect]
  );

  // 树节点展开/收起事件处理
  const handleExpand = useCallback((expandedKeys: React.Key[]) => {
    setExpandedKeys(expandedKeys);
    setAutoExpandParent(false);
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

  // 刷新树数据
  const refreshTree = useCallback(() => {
    loadTreeData();
  }, [loadTreeData]);

  // 暴露方法给父组件
  useImperativeHandle(
    ref,
    () => ({
      expandAll,
      collapseAll,
      refresh: refreshTree,
      getLoading: () => loading,
    }),
    [expandAll, collapseAll, refreshTree, loading]
  );

  // 初始化加载数据
  useEffect(() => {
    loadTreeData();
  }, []);

  // 同步外部选中的组织ID
  useEffect(() => {
    if (selectedOrgId) {
      setSelectedKeys([selectedOrgId]);
    }
  }, [selectedOrgId]);

  return (
    <div className={styles.container}>
      {/* 工具栏（可选） */}
      {!hideToolbar && (
        <div className={styles.toolbar}>
          <Button
            size="small"
            icon={<ExpandAltOutlined />}
            onClick={expandAll}
            title="全部展开"
          />
          <Button
            size="small"
            icon={<CompressOutlined />}
            onClick={collapseAll}
            title="全部收起"
          />
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={refreshTree}
            loading={loading}
            title="刷新"
          />
        </div>
      )}

      {/* 树内容区域 */}
      <div
        className={
          hideToolbar ? styles.treeContentNoToolbar : styles.treeContent
        }
      >
        <Spin spinning={loading}>
          {treeData.length === 0 ? (
            <Empty
              description="暂无组织数据"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          ) : (
            <Tree
              treeData={treeData}
              selectedKeys={selectedKeys}
              expandedKeys={expandedKeys}
              autoExpandParent={autoExpandParent}
              onSelect={handleSelect}
              onExpand={handleExpand}
              showLine={{ showLeafIcon: false }}
              blockNode
            />
          )}
        </Spin>
      </div>
    </div>
  );
});

OrganizationSelector.displayName = "OrganizationSelector";

export default OrganizationSelector;
