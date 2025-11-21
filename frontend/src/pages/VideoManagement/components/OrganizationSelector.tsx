import { getOrganizationTree } from "@/services/organization";
import { AppstoreOutlined, UnorderedListOutlined } from "@ant-design/icons";
import { Button, message, Select, Space, TreeSelect } from "antd";
import React, { useEffect, useState } from "react";

interface OrganizationSelectorProps {
  value?: string;
  onChange?: (value: string | undefined) => void;
  placeholder?: string;
  style?: React.CSSProperties;
  allowClear?: boolean;
  mode?: "list" | "tree"; // 默认使用列表模式
}

/**
 * 组织架构选择器 - 改进版
 * 支持两种模式：
 * 1. 列表模式（默认）：扁平化列表，适合快速搜索和选择
 * 2. 树形模式：层级展示，适合了解组织结构
 */
const OrganizationSelector: React.FC<OrganizationSelectorProps> = ({
  value,
  onChange,
  placeholder = "选择组织架构",
  style,
  allowClear = true,
  mode = "list",
}) => {
  const [organizationTreeData, setOrganizationTreeData] = useState<any[]>([]);
  const [organizationListData, setOrganizationListData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [viewMode, setViewMode] = useState<"list" | "tree">(mode);

  // 转换组织树数据为TreeSelect格式
  const convertToTreeData = (orgs: API.OrganizationTree[]): any[] => {
    return orgs.map((org) => ({
      title: org.name,
      value: org.id,
      key: org.id,
      children: org.children ? convertToTreeData(org.children) : undefined,
    }));
  };

  // 将组织树扁平化为列表，保留完整路径信息
  const flattenOrgTree = (
    orgs: API.OrganizationTree[],
    parentPath: string = ""
  ): Array<{ label: string; value: string; path: string }> => {
    let result: Array<{ label: string; value: string; path: string }> = [];

    orgs.forEach((org) => {
      const currentPath = parentPath ? `${parentPath} > ${org.name}` : org.name;

      result.push({
        label: currentPath,
        value: org.id,
        path: currentPath,
      });

      if (org.children && org.children.length > 0) {
        result = result.concat(flattenOrgTree(org.children, currentPath));
      }
    });

    return result;
  };

  // 加载组织树数据
  const loadOrganizationTree = async () => {
    setLoading(true);
    try {
      const response = await getOrganizationTree({});
      if (response.code === 0 && response.data) {
        // 生成树形数据
        const treeData = convertToTreeData(response.data);
        setOrganizationTreeData(treeData);

        // 生成扁平列表数据
        const listData = flattenOrgTree(response.data);
        setOrganizationListData(listData);
      } else {
        message.error(response.msg || "获取组织架构失败");
      }
    } catch (error) {
      console.error("获取组织架构失败:", error);
      message.error("获取组织架构失败");
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载
  useEffect(() => {
    loadOrganizationTree();
  }, []);

  // 列表模式渲染
  const renderListMode = () => (
    <Select
      style={{ minWidth: 200, ...style }}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      allowClear={allowClear}
      loading={loading}
      showSearch
      optionFilterProp="label"
      options={organizationListData}
      filterOption={(input, option) =>
        (option?.label ?? "").toLowerCase().includes(input.toLowerCase())
      }
    />
  );

  // 树形模式渲染
  const renderTreeMode = () => (
    <TreeSelect
      style={{ minWidth: 200, ...style }}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      allowClear={allowClear}
      loading={loading}
      treeData={organizationTreeData}
      treeDefaultExpandAll={false}
      showSearch
      treeNodeFilterProp="title"
      styles={{ popup: { root: { maxHeight: 400, overflow: "auto" } } }}
    />
  );

  return (
    <Space.Compact style={{ width: "100%" }}>
      {viewMode === "list" ? renderListMode() : renderTreeMode()}
      <Button
        icon={
          viewMode === "list" ? <AppstoreOutlined /> : <UnorderedListOutlined />
        }
        onClick={() => setViewMode(viewMode === "list" ? "tree" : "list")}
        title={viewMode === "list" ? "切换到树形模式" : "切换到列表模式"}
      />
    </Space.Compact>
  );
};

export default OrganizationSelector;
