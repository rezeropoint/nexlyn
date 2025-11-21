import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import CreateTagForm from "./components/CreateTagForm";
import EditTagForm from "./components/EditTagForm";
import TagTable from "./components/TagTable";
import { useDeviceTag } from "./hooks/useDeviceTag";

/**
 * 设备标签管理主页面
 *
 * 组件结构：
 * - 主页面负责组件编排和状态协调
 * - 业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const DeviceTagManagement: React.FC = () => {
  const actionRef = useRef<ActionType>();

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,
    selectionState,

    // 设置方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleBatchDelete,
    handleEditClick,
  } = useDeviceTag();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建标签成功后的回调
  const handleCreateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新标签成功后的回调
  const handleUpdateSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 标签操作处理（包含表格刷新）
  const handleTagOperation = {
    delete: async (record: any) => {
      await handleDelete(record);
      handleTableRefresh();
    },
    batchDelete: async (tags: any[]) => {
      await handleBatchDelete(tags);
      handleTableRefresh();
    },
  };

  return (
    <PageContainer subTitle="定义和管理物联网设备的标签分类，支持标签颜色自定义">
      {/* 主表格 */}
      <TagTable
        actionRef={actionRef}
        selectionState={selectionState}
        onSelectionChange={(state) =>
          setSelectionState((prev) => ({ ...prev, ...state }))
        }
        onCreateTag={() => setModalOpen("create", true)}
        onEditTag={handleEditClick}
        onDeleteTag={handleTagOperation.delete}
        onBatchDelete={handleTagOperation.batchDelete}
      />

      {/* 创建标签表单 */}
      <CreateTagForm
        open={modalState.create}
        onOpenChange={(open) => setModalOpen("create", open)}
        onFinish={handleCreateSuccess}
      />

      {/* 编辑标签表单 */}
      <EditTagForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />
    </PageContainer>
  );
};

export default DeviceTagManagement;
