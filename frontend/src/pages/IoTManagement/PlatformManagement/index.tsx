import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import CreatePlatformForm from "./components/CreatePlatformForm";
import EditPlatformForm from "./components/EditPlatformForm";
import PlatformTable from "./components/PlatformTable";
import { usePlatform } from "./hooks/usePlatform";

/**
 * 平台管理主页面
 *
 * 组件结构：
 * - 主页面负责组件编排和状态协调
 * - 业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const PlatformManagement: React.FC = () => {
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
    handleEditClick,
  } = usePlatform();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建平台配置成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新平台配置成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 删除平台配置后刷新表格
  const handleDeletePlatform = async (record: any) => {
    await handleDelete(record);
    handleTableRefresh();
  };

  return (
    <PageContainer subTitle="配置IoT数据分发目标平台，支持Skylark流程引擎、Webhook和Kafka集成">
      {/* 主表格 */}
      <PlatformTable
        actionRef={actionRef}
        onCreatePlatform={() => setModalOpen("create", true)}
        onEditPlatform={handleEditClick}
        onDeletePlatform={handleDeletePlatform}
      />

      {/* 创建平台配置表单 */}
      <CreatePlatformForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen("create", open)}
        onFinish={handleCreateSuccess}
      />

      {/* 编辑平台配置表单 */}
      <EditPlatformForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />
    </PageContainer>
  );
};

export default PlatformManagement;
