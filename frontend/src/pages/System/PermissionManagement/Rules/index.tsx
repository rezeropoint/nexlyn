import { CheckCircleOutlined } from "@ant-design/icons";
import { type ActionType, PageContainer } from "@ant-design/pro-components";
import { useAccess } from "@umijs/max";
import { Alert } from "antd";
import React, { useRef } from "react";
import RulesTable from "./components/RulesTable";
import { useRulesManagement } from "./hooks/useRulesManagement";
import styles from "./index.less";

const PermissionRules: React.FC = () => {
  const access = useAccess();
  const actionRef = useRef<ActionType>();
  const { handleAddRule, handleDeleteRule, handleCheckPermission } =
    useRulesManagement();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 添加权限规则成功后的回调
  const handleAddRuleSuccess = async (
    values: API.AddPermissionRequest
  ): Promise<boolean> => {
    const success = await handleAddRule(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 删除权限规则后的回调
  const handleDeleteRuleWithRefresh = async (
    record: API.PermissionRule
  ): Promise<void> => {
    await handleDeleteRule(record);
    handleTableRefresh();
  };

  // 权限检查
  if (!access.canReadPermissions?.()) {
    return (
      <PageContainer>
        <div className={styles.noPermissionContainer}>
          <CheckCircleOutlined className={styles.noPermissionIcon} />
          <h2>权限不足</h2>
          <p>您需要权限管理权限才能访问此页面</p>
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Alert
        message="📋 功能说明"
        description={
          <div>
            • <strong>用户直接权限管理</strong>
            ：为特定用户添加额外的权限规则（在角色权限基础之上）
            <br />• <strong>角色权限管理</strong>
            ：所有角色权限请通过"角色权限管理"页面统一配置
            <br />• <strong>权限叠加</strong>：用户最终权限 = 角色权限 +
            直接权限
            <br />• 此页面只显示直接分配给用户的权限，不包含从角色继承的权限
          </div>
        }
        type="success"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <RulesTable
        actionRef={actionRef}
        onCheckPermission={handleCheckPermission}
        onDeleteRule={handleDeleteRuleWithRefresh}
        onAddRule={handleAddRuleSuccess}
      />
    </PageContainer>
  );
};

export default PermissionRules;
