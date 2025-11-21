import { type ActionType, PageContainer } from "@ant-design/pro-components";
import React, { useRef } from "react";
import OrgMappingForm from "./components/OrgMappingForm";
import OrgMappingTable from "./components/OrgMappingTable";
import { useOrgMapping } from "./hooks/useOrgMapping";

const OrgMapping: React.FC = () => {
  const actionRef = useRef<ActionType>();

  const {
    currentRow,
    modalState,
    setModalOpen,
    handleCreate,
    handleUpdate,
    handleDelete,
    handleEditClick,
  } = useOrgMapping();

  const handleTableRefresh = () => {
    actionRef.current?.reload();
  };

  const handleFinish = async (values: any) => {
    const success = modalState.edit
      ? await handleUpdate(values)
      : await handleCreate(values);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  const handleDeleteMapping = async (record: any) => {
    const success = await handleDelete(record);
    if (success) {
      handleTableRefresh();
    }
    return success;
  };

  return (
    <PageContainer subTitle="将Skylark流程中的组织字段值映射到本地组织架构，实现权限过滤">
      <OrgMappingTable
        actionRef={actionRef}
        onCreate={() => setModalOpen("create", true)}
        onEdit={handleEditClick}
        onDelete={handleDeleteMapping}
      />
      <OrgMappingForm
        open={modalState.create || modalState.edit}
        onOpenChange={(open) => {
          // 使用回调形式，避免闭包拿到过期的 modalState
          setModalOpen("create", open && !modalState.edit);
          setModalOpen("edit", open && !!modalState.edit);
        }}
        onFinish={handleFinish}
        currentRow={currentRow}
        isEditMode={modalState.edit}
      />
    </PageContainer>
  );
};

export default OrgMapping;
