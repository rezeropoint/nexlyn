import {
  CompressOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  ExpandAltOutlined,
  PlusOutlined,
  ReloadOutlined,
  SwapOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { useAccess } from "@umijs/max";
import { Button, Dropdown, Empty, Modal, Spin, Tag, Tooltip, Tree } from "antd";
import React, { useMemo } from "react";
import {
  ORGANIZATION_STATUS_COLORS,
  ORGANIZATION_STATUS_TEXT,
} from "../constants";
import type { TreeDropInfo, TreeNodeData, TreeSelectInfo } from "../types";
import styles from "./OrganizationTree.less";

const { confirm } = Modal;

interface OrganizationTreeProps {
  treeData: TreeNodeData[];
  expandedKeys: React.Key[];
  selectedKeys: React.Key[];
  autoExpandParent: boolean;
  loading: boolean;
  onSelect: (selectedKeys: React.Key[], info: TreeSelectInfo) => void;
  onExpand: (expandedKeys: React.Key[]) => void;
  onDragStart?: (info: any) => void;
  onDragEnter?: (info: any) => void;
  onDrop?: (info: TreeDropInfo) => void;
  onRefresh: () => void;
  onCreateOrganization: (parentId?: string) => void;
  onEditOrganization: (
    record: API.Organization | API.OrganizationBrief
  ) => void;
  onDeleteOrganization: (
    record: API.Organization | API.OrganizationBrief
  ) => void;
  onViewMembers: (record: API.Organization | API.OrganizationBrief) => void;
  onMoveOrganization: (
    record: API.Organization | API.OrganizationBrief
  ) => void;
  onExpandAll: () => void;
  onCollapseAll: () => void;
}

/**
 * 组织树展示组件
 */
const OrganizationTree: React.FC<OrganizationTreeProps> = ({
  treeData,
  expandedKeys,
  selectedKeys,
  autoExpandParent,
  loading,
  onSelect,
  onExpand,
  onDragStart,
  onDragEnter,
  onDrop,
  onRefresh,
  onCreateOrganization,
  onEditOrganization,
  onDeleteOrganization,
  onViewMembers,
  onMoveOrganization,
  onExpandAll,
  onCollapseAll,
}) => {
  const access = useAccess();

  // 处理删除确认
  const handleDeleteClick = (record: TreeNodeData, e: React.MouseEvent) => {
    e.stopPropagation();
    confirm({
      title: "确认删除",
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>
            确定要删除组织 <strong>{record.name}</strong> 吗？
          </p>
          {record.children && record.children.length > 0 && (
            <p className={styles.deleteWarning}>
              注意：该组织包含 {record.children.length}{" "}
              个子组织，删除后子组织也会被删除。
            </p>
          )}
          <p className={styles.deleteHint}>此操作不可恢复，请谨慎操作。</p>
        </div>
      ),
      okText: "确认删除",
      okType: "danger",
      cancelText: "取消",
      onOk: () => onDeleteOrganization(record),
    });
  };

  // 右键菜单配置
  const getContextMenuItems = (node: TreeNodeData) => [
    {
      key: "create",
      label: "新增子组织",
      icon: <PlusOutlined />,
      disabled: !access.canCreateOrganizations,
      onClick: () => onCreateOrganization(node.id),
    },
    {
      type: "divider" as const,
    },
    {
      key: "edit",
      label: "编辑组织",
      icon: <EditOutlined />,
      disabled: !access.canUpdateOrganizations,
      onClick: () => onEditOrganization(node),
    },
    {
      key: "move",
      label: "移动组织",
      icon: <SwapOutlined />,
      disabled: !access.canMoveOrganizations,
      onClick: () => onMoveOrganization(node),
    },
    {
      key: "members",
      label: "查看成员",
      icon: <TeamOutlined />,
      onClick: () => onViewMembers(node),
    },
    {
      type: "divider" as const,
    },
    {
      key: "delete",
      label: "删除组织",
      icon: <DeleteOutlined />,
      danger: true,
      disabled: !access.canDeleteOrganizations,
      onClick: (e: any) => handleDeleteClick(node, e.domEvent),
    },
  ];

  // 自定义树节点渲染
  const renderTreeNode = (node: TreeNodeData) => {
    return (
      <Dropdown
        menu={{ items: getContextMenuItems(node) }}
        trigger={["contextMenu"]}
      >
        <div className={styles.customTreeNode}>
          <div className={styles.nodeMain}>
            <span className={styles.nodeName}>{node.name}</span>
            <span className={styles.nodeCode}>({node.code})</span>
            <Tag
              color={
                ORGANIZATION_STATUS_COLORS[
                  node.status as keyof typeof ORGANIZATION_STATUS_COLORS
                ]
              }
              className={styles.nodeTag}
            >
              {
                ORGANIZATION_STATUS_TEXT[
                  node.status as keyof typeof ORGANIZATION_STATUS_TEXT
                ]
              }
            </Tag>
          </div>
          <div className={styles.nodeInfo}>
            {node.managerName && (
              <Tooltip title={`负责人: ${node.managerName}`}>
                <span className={styles.managerInfo}>
                  👤 {node.managerName}
                </span>
              </Tooltip>
            )}
            <Tooltip title="成员数量">
              <span className={styles.memberCount}>👥 {node.memberCount}</span>
            </Tooltip>
          </div>
        </div>
      </Dropdown>
    );
  };

  // 处理树数据，添加自定义渲染
  const processedTreeData = useMemo(() => {
    const processNode = (node: TreeNodeData): TreeNodeData => ({
      ...node,
      title: renderTreeNode(node),
      children: node.children?.map(processNode),
    });

    return treeData.map(processNode);
  }, [treeData, access]);

  return (
    <div className={styles.treeContainer}>
      {/* 工具栏 */}
      <div className={styles.toolbar}>
        <Button size="small" icon={<ExpandAltOutlined />} onClick={onExpandAll}>
          全部展开
        </Button>
        <Button
          size="small"
          icon={<CompressOutlined />}
          onClick={onCollapseAll}
        >
          全部收起
        </Button>
        <Button
          size="small"
          icon={<ReloadOutlined />}
          onClick={onRefresh}
          loading={loading}
        >
          刷新
        </Button>
      </div>

      {/* 树内容区域 */}
      <div className={styles.treeContent}>
        <Spin spinning={loading}>
          {treeData.length === 0 ? (
            <Empty
              description="暂无组织数据"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          ) : (
            <Tree
              treeData={processedTreeData}
              selectedKeys={selectedKeys}
              expandedKeys={expandedKeys}
              autoExpandParent={autoExpandParent}
              onSelect={onSelect}
              onExpand={onExpand}
              onDragStart={onDragStart}
              onDragEnter={onDragEnter}
              onDrop={onDrop}
              draggable={access.canMoveOrganizations}
              showLine={{ showLeafIcon: false }}
              height={400}
              virtual={treeData.length > 100}
              blockNode
            />
          )}
        </Spin>
      </div>
    </div>
  );
};

export default OrganizationTree;
