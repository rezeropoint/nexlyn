import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import CreateTemplateForm from "./components/CreateTemplateForm";
import EditTemplateForm from "./components/EditTemplateForm";
import TemplateTable from "./components/TemplateTable";
import { useSensorTemplate } from "./hooks/useSensorTemplate";

/**
 * 设备模板管理主页面
 *
 * 组件结构：
 * - 主页面负责组件编排和状态协调
 * - 业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const SensorTemplateManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,

    // 设置方法
    setModalOpen,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleToggleEnabled,
    handleEditClick,
  } = useSensorTemplate();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建模板成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新模板成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 模板操作处理（包含表格刷新）
  const handleTemplateOperation = {
    delete: async (record: any) => {
      await handleDelete(record);
      handleTableRefresh();
    },
    toggleEnabled: async (record: any) => {
      await handleToggleEnabled(record);
      handleTableRefresh();
    },
  };

  return (
    <PageContainer subTitle="配置物联网设备的MQTT数据解析模板、字段映射及业务数据处理规则">
      {/* 主表格 */}
      <TemplateTable
        actionRef={actionRef}
        onCreateTemplate={() => setModalOpen("create", true)}
        onEditTemplate={handleEditClick}
        onDeleteTemplate={handleTemplateOperation.delete}
      />

      {/* 创建模板表单 */}
      <CreateTemplateForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen("create", open)}
        onFinish={handleCreateSuccess}
      />

      {/* 编辑模板表单 */}
      <EditTemplateForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />
    </PageContainer>
  );
};

export default SensorTemplateManagement;
