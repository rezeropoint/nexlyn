import {
  CheckCircleOutlined,
  InfoCircleOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { Tabs } from "antd";
import React, { useState } from "react";
import PermissionChecker from "./components/PermissionChecker";
import PermissionGuide from "./components/PermissionGuide";
import RoleGuide from "./components/RoleGuide";

/**
 * 权限检查工具页面
 *
 * 主要功能：
 * 1. 权限检查工具 - 检查指定用户在指定租户下的权限
 * 2. 权限说明 - 显示权限相关的指导信息
 * 3. 角色说明 - 显示角色相关的指导信息
 */
const PermissionCheckerPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('checker');

  // Tab配置
  const tabItems = [
    {
      key: 'checker',
      label: (
        <span>
          <CheckCircleOutlined />
          权限检查工具
        </span>
      ),
      children: <PermissionChecker />,
    },
    {
      key: 'permission-guide',
      label: (
        <span>
          <InfoCircleOutlined />
          权限说明
        </span>
      ),
      children: <PermissionGuide />,
    },
    {
      key: 'role-guide',
      label: (
        <span>
          <TeamOutlined />
          角色说明
        </span>
      ),
      children: <RoleGuide />,
    },
  ];

  return (
    <PageContainer
      title="权限检查工具"
      subTitle="检查用户权限、查看权限和角色说明"
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        type="card"
        style={{ padding: '0 24px' }}
      />
    </PageContainer>
  );
};

export default PermissionCheckerPage;
