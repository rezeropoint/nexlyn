import type { PermissionItem } from "@/services/permission/role";
import { SearchOutlined } from "@ant-design/icons";
import { Card, Empty, Flex, Input, Tag, Tree, Typography } from "antd";
import type { DataNode } from "antd/es/tree";
import React, { useCallback, useMemo, useState } from "react";
import styles from "./PermissionSelector.less";

const { Text } = Typography;

interface PermissionSelectorProps {
  value?: string[];
  onChange?: (value: string[]) => void;
  availablePermissions: PermissionItem[];
  disabled?: boolean;
}

// 前缀常量，用于区分不同层级的节点
const CATEGORY_PREFIX = "category:";
const RESOURCE_PREFIX = "resource:";

const PermissionSelector: React.FC<PermissionSelectorProps> = ({
  value = [],
  onChange,
  availablePermissions,
  disabled = false,
}) => {
  const [searchValue, setSearchValue] = useState("");
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [autoExpandParent, setAutoExpandParent] = useState(true);

  // 按分类和资源分组权限：category -> resource -> permissions
  const groupedPermissions = useMemo(() => {
    const result: Record<string, Record<string, PermissionItem[]>> = {};

    for (const item of availablePermissions) {
      if (!result[item.category]) {
        result[item.category] = {};
      }
      if (!result[item.category][item.resource]) {
        result[item.category][item.resource] = [];
      }
      result[item.category][item.resource].push(item);
    }

    return result;
  }, [availablePermissions]);

  // 分类排序顺序
  const categoryOrder = useMemo(() => {
    return Object.keys(groupedPermissions).sort((a, b) => a.localeCompare(b));
  }, [groupedPermissions]);

  // 构建三级树形数据：分类 -> 资源 -> 权限
  const treeData = useMemo((): DataNode[] => {
    return categoryOrder.map((category) => {
      const resources = groupedPermissions[category];
      const resourceKeys = Object.keys(resources).sort((a, b) =>
        a.localeCompare(b)
      );

      // 统计该分类下的总权限数
      const totalPerms = resourceKeys.reduce(
        (sum, r) => sum + resources[r].length,
        0
      );

      return {
        key: `${CATEGORY_PREFIX}${category}`,
        title: (
          <Flex align="center" gap={8}>
            <Text strong>{category}</Text>
            <Tag className={styles.countTag}>{totalPerms}</Tag>
          </Flex>
        ),
        children: resourceKeys.map((resource) => {
          const permissions = resources[resource];
          // 使用后端返回的 resourceName
          const resourceName = permissions[0]?.resourceName || resource;

          return {
            key: `${RESOURCE_PREFIX}${category}:${resource}`,
            title: (
              <Flex align="center" gap={8}>
                <Text>{resourceName}</Text>
                <Tag className={styles.countTag}>{permissions.length}</Tag>
              </Flex>
            ),
            children: permissions.map((perm) => ({
              key: `${perm.resource}:${perm.action}`,
              title: (
                <Flex align="center" gap={8} className={styles.permissionItem}>
                  <Text>{perm.description}</Text>
                  <Text type="secondary" className={styles.permissionKey}>
                    {perm.action}
                  </Text>
                </Flex>
              ),
            })),
          };
        }),
      };
    });
  }, [categoryOrder, groupedPermissions]);

  // 搜索过滤后的树形数据
  const filteredTreeData = useMemo((): DataNode[] => {
    if (!searchValue.trim()) {
      return treeData;
    }

    const lowerSearch = searchValue.toLowerCase();

    return treeData
      .map((categoryNode) => {
        const category = (categoryNode.key as string).replace(
          CATEGORY_PREFIX,
          ""
        );
        const categoryMatches = category.toLowerCase().includes(lowerSearch);

        // 过滤资源节点
        const filteredResources = (categoryNode.children || [])
          .map((resourceNode) => {
            const resourceKey = (resourceNode.key as string).replace(
              RESOURCE_PREFIX,
              ""
            );
            const resource = resourceKey.split(":")[1];

            // 过滤权限节点
            const filteredPerms = (resourceNode.children || []).filter(
              (permNode) => {
                const permKey = permNode.key as string;
                const perm = availablePermissions.find(
                  (p) => `${p.resource}:${p.action}` === permKey
                );
                if (!perm) return false;

                return (
                  categoryMatches ||
                  perm.description.toLowerCase().includes(lowerSearch) ||
                  perm.resourceName.toLowerCase().includes(lowerSearch) ||
                  perm.action.toLowerCase().includes(lowerSearch)
                );
              }
            );

            if (filteredPerms.length === 0) {
              return null;
            }

            return {
              ...resourceNode,
              children: filteredPerms,
            };
          })
          .filter(Boolean) as DataNode[];

        if (filteredResources.length === 0) {
          return null;
        }

        return {
          ...categoryNode,
          children: filteredResources,
        };
      })
      .filter(Boolean) as DataNode[];
  }, [treeData, searchValue, availablePermissions]);

  // 搜索时自动展开所有匹配的节点
  const searchExpandedKeys = useMemo(() => {
    if (!searchValue.trim()) {
      return expandedKeys;
    }

    const keys: React.Key[] = [];
    for (const categoryNode of filteredTreeData) {
      keys.push(categoryNode.key);
      for (const resourceNode of categoryNode.children || []) {
        keys.push(resourceNode.key);
      }
    }
    return keys;
  }, [searchValue, filteredTreeData, expandedKeys]);

  // 处理勾选变化
  const handleCheck = useCallback(
    (
      checked: React.Key[] | { checked: React.Key[]; halfChecked: React.Key[] }
    ) => {
      const checkedKeys = Array.isArray(checked) ? checked : checked.checked;

      // 过滤掉分类和资源节点的 key，只保留权限节点的 key
      const permissionKeys = checkedKeys
        .map(String)
        .filter(
          (key) =>
            !key.startsWith(CATEGORY_PREFIX) &&
            !key.startsWith(RESOURCE_PREFIX)
        );

      onChange?.(permissionKeys);
    },
    [onChange]
  );

  // 处理展开/折叠
  const handleExpand = useCallback((keys: React.Key[]) => {
    setExpandedKeys(keys);
    setAutoExpandParent(false);
  }, []);

  // 统计已选择的权限（按分类）
  const selectedStats = useMemo(() => {
    const stats: Record<string, { selected: number; total: number }> = {};

    for (const category of categoryOrder) {
      const resources = groupedPermissions[category];
      let selected = 0;
      let total = 0;

      for (const resource of Object.keys(resources)) {
        const permissions = resources[resource];
        total += permissions.length;
        selected += permissions.filter((perm) =>
          value.includes(`${perm.resource}:${perm.action}`)
        ).length;
      }

      stats[category] = { selected, total };
    }

    return stats;
  }, [categoryOrder, groupedPermissions, value]);

  // 全选/全不选
  const handleSelectAll = useCallback(() => {
    if (value.length === availablePermissions.length) {
      onChange?.([]);
    } else {
      const allKeys = availablePermissions.map(
        (p) => `${p.resource}:${p.action}`
      );
      onChange?.(allKeys);
    }
  }, [value.length, availablePermissions, onChange]);

  if (availablePermissions.length === 0) {
    return (
      <Card size="small">
        <Empty description="暂无可用权限" />
      </Card>
    );
  }

  return (
    <Flex vertical gap={12}>
      {/* 搜索框和统计 */}
      <Flex justify="space-between" align="center" gap={16}>
        <Input
          placeholder="搜索权限（支持描述、资源名、分类）"
          prefix={<SearchOutlined className={styles.searchIcon} />}
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          allowClear
          disabled={disabled}
          className={styles.searchInput}
        />
        <Flex align="center" gap={8}>
          <Text type="secondary">
            已选 {value.length}/{availablePermissions.length}
          </Text>
          {!disabled && (
            <Typography.Link onClick={handleSelectAll}>
              {value.length === availablePermissions.length
                ? "取消全选"
                : "全选"}
            </Typography.Link>
          )}
        </Flex>
      </Flex>

      {/* 分类统计 */}
      <Card size="small" className={styles.statsCard}>
        <Flex wrap="wrap" gap={4}>
          {categoryOrder.map((category) => {
            const stat = selectedStats[category];
            const isFullSelected = stat.selected === stat.total;
            const isPartialSelected = stat.selected > 0 && !isFullSelected;

            return (
              <Tag
                key={category}
                color={
                  isFullSelected
                    ? "success"
                    : isPartialSelected
                      ? "processing"
                      : "default"
                }
              >
                {category}: {stat.selected}/{stat.total}
              </Tag>
            );
          })}
        </Flex>
      </Card>

      {/* 权限树 */}
      <Card size="small" className={styles.treeCard}>
        {filteredTreeData.length > 0 ? (
          <Tree
            checkable
            disabled={disabled}
            checkedKeys={value}
            expandedKeys={
              searchValue.trim() ? searchExpandedKeys : expandedKeys
            }
            autoExpandParent={autoExpandParent}
            onCheck={handleCheck}
            onExpand={handleExpand}
            treeData={filteredTreeData}
            className={styles.permissionTree}
            selectable={false}
          />
        ) : (
          <Empty description="未找到匹配的权限" />
        )}
      </Card>
    </Flex>
  );
};

export default PermissionSelector;
