import { TreeSelect } from "antd";
import React from "react";
import type { MockOrganization } from "../mockData/organizationMock";
import { mockOrganizations } from "../mockData/organizationMock";

interface OrganizationFilterProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  style?: React.CSSProperties;
  allowClear?: boolean;
}

const OrganizationFilter: React.FC<OrganizationFilterProps> = ({
  value,
  onChange,
  placeholder = "选择组织架构",
  style,
  allowClear = true,
}) => {
  // 转换数据为TreeSelect需要的格式
  const convertToTreeData = (orgs: MockOrganization[]): any[] => {
    return orgs.map((org) => ({
      title: org.name,
      value: org.id,
      key: org.id,
      children: org.children ? convertToTreeData(org.children) : undefined,
    }));
  };

  const treeData = convertToTreeData(mockOrganizations);

  return (
    <TreeSelect
      style={{ minWidth: 200, ...style }}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      allowClear={allowClear}
      treeData={treeData}
      treeDefaultExpandAll={false}
      treeDefaultExpandedKeys={["org-china"]}
      showSearch
      treeNodeFilterProp="title"
      styles={{ popup: { root: { maxHeight: 400, overflow: "auto" } } }}
    />
  );
};

export default OrganizationFilter;
