import { ExclamationCircleOutlined } from "@ant-design/icons";
import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import CreateTagForm from "./components/CreateTagForm";
import EditTagForm from "./components/EditTagForm";
import TagTable from "./components/TagTable";
import { useTagManagement } from "./hooks/useTagManagement";
import { useTagPermissions } from "./hooks/useTagPermissions";
import styles from "./index.less";

/**
 * 标签管理主页面
 *
 * 重构后的组件结构：
 * - 主页面只负责组件编排和状态协调
 * - 具体业务逻辑由自定义Hook管理
 * - UI组件拆分为可复用的子组件
 */
const TagManagement: React.FC = () => {
  const permissions = useTagPermissions();
  const actionRef = useRef<ActionType>();

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,
    selectionState,
    defaultScope,

    // 操作方法
    setModalOpen,
    setSelectionState,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleBatchDelete,
    handleEditClick,
  } = useTagManagement();

  // 处理表格刷新
  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  // 创建标签成功后的回调
  const handleCreateSuccess = async (
    values: API.CreateTagRequest
  ): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 更新标签成功后的回调
  const handleUpdateSuccess = async (
    values: API.UpdateTagRequest
  ): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  // 标签操作处理（包含表格刷新）
  const handleTagOperation = {
    delete: async (record: API.TagDefinition) => {
      await handleDelete(record);
      handleTableRefresh();
    },
    batchDelete: async (tags: API.TagDefinition[]) => {
      await handleBatchDelete(tags);
      handleTableRefresh();
    },
  };

  // 权限检查
  if (!permissions.canRead) {
    return (
      <PageContainer>
        <div className={styles.noPermissionContainer}>
          <ExclamationCircleOutlined className={styles.noPermissionIcon} />
          <h2>权限不足</h2>
          <p>您需要标签管理权限才能访问此页面</p>
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer subTitle="定义和管理系统级别的标签分类，支持租户级和全局级标签">
      {/* 主表格 */}
      <TagTable
        actionRef={actionRef}
        selectionState={selectionState}
        onSelectionChange={setSelectionState}
        onCreateTag={() => setModalOpen("create", true)}
        onEditTag={handleEditClick}
        onDeleteTag={handleTagOperation.delete}
        onBatchDelete={handleTagOperation.batchDelete}
        defaultScope={defaultScope}
      />

      {/* 创建标签表单 */}
      {permissions.canCreate && (
        <CreateTagForm
          open={modalState.create}
          onOpenChange={(open) => setModalOpen("create", open)}
          onFinish={handleCreateSuccess}
        />
      )}

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

export default TagManagement;
