import {
  type ActionType,
  PageContainer,
  ProCard,
} from "@ant-design/pro-components";
import { App, Modal, Splitter } from "antd";
import React, { useRef } from "react";
import useResponsiveLayout from "@/hooks/useResponsiveLayout";
import { syncOrganization, unbindOrganization } from "@/services/sync";
import BindOrganizationForm from "./components/BindOrganizationForm";
import CreateOrganizationForm from "./components/CreateOrganizationForm";
import EditOrganizationForm from "./components/EditOrganizationForm";
import MoveOrganizationModal from "./components/MoveOrganizationModal";
import OrganizationMembers from "./components/OrganizationMembers";
import OrganizationTable from "./components/OrganizationTable";
import OrganizationTree from "./components/OrganizationTree";
import { useOrganizationManagement } from "./hooks/useOrganizationManagement";
import { useOrganizationTree } from "./hooks/useOrganizationTree";
import styles from "./index.less";

/**
 * 组织管理主页面
 *
 * 功能特性：
 * - 左右分栏布局，左侧组织树，右侧组织列表
 * - 组织CRUD操作（创建、编辑、删除）
 * - 组织树拖拽移动功能
 * - 组织成员管理（添加/移除/修改成员关系）
 * - 批量操作支持
 * - 权限控制（基于organization权限）
 * - 租户隔离（超级管理员可跨租户管理）
 * - 树表联动，选中树节点自动过滤表格
 */
const OrganizationManagement: React.FC = () => {
  const tableActionRef = useRef<ActionType>();
  const layout = useResponsiveLayout({ breakpoint: "xl" });
  const { message } = App.useApp();
  const [createParentId, setCreateParentId] = React.useState<
    string | undefined
  >(undefined);

  // 使用自定义Hook管理所有业务逻辑
  const {
    // 状态
    currentRow,
    modalState,
    selectionState,
    selectedOrgId,

    // 操作方法
    setModalOpen,
    setSelectionState,
    setCurrentRow,

    // 业务操作
    handleCreate,
    handleUpdate,
    handleDelete,
    handleMove,
    handleBatchDelete,
    handleToggleStatus,
    handleBatchToggleStatus,
    handleEditClick,
    handleMembersClick,
    handleMoveClick,
  } = useOrganizationManagement();

  // 组织树相关逻辑
  const {
    treeData,
    expandedKeys,
    selectedKeys,
    autoExpandParent,
    loading: treeLoading,
    onSelect,
    onExpand,
    onDragStart,
    onDragEnter,
    onDrop,
    expandAll,
    collapseAll,
    refreshTree,
  } = useOrganizationTree();

  // 处理树节点选择，更新表格过滤
  const handleTreeSelect = (selectedKeys: React.Key[], info: any) => {
    onSelect(selectedKeys, info);

    // 更新表格过滤的父组织ID
    if (selectedKeys.length > 0) {
      const selectedNodeId = selectedKeys[0] as string;
      // 通过设置selectedOrgId来触发表格刷新，显示选中组织的子组织
      if (selectedNodeId !== selectedOrgId) {
        tableActionRef.current?.reload();
      }
    } else {
      // 如果没有选中节点，清空过滤
      if (selectedOrgId) {
        tableActionRef.current?.reload();
      }
    }
  };

  // 获取当前选中的组织ID，用于表格过滤
  const getCurrentSelectedOrgId = () => {
    return selectedKeys.length > 0 ? (selectedKeys[0] as string) : "";
  };

  // 处理表格刷新
  const handleTableRefresh = () => {
    tableActionRef.current?.reload();
  };

  // 处理树刷新
  const handleTreeRefresh = () => {
    refreshTree();
  };

  // 创建组织成功后的回调
  const handleCreateSuccess = async (
    values: API.CreateOrganizationRequest
  ): Promise<boolean> => {
    const success = await handleCreate(values);
    if (success) {
      handleTreeRefresh();
      handleTableRefresh();
    }
    return success;
  };

  // 更新组织成功后的回调
  const handleUpdateSuccess = async (
    values: API.UpdateOrganizationRequest
  ): Promise<boolean> => {
    const success = await handleUpdate(values);
    if (success) {
      handleTreeRefresh();
      handleTableRefresh();
    }
    return success;
  };

  // 移动组织成功后的回调
  const handleMoveSuccess = async (values: {
    newParentId: string;
    newSortOrder?: number;
  }): Promise<boolean> => {
    const success = await handleMove(values);
    if (success) {
      handleTreeRefresh();
      handleTableRefresh();
    }
    return success;
  };

  // 组织操作处理（包含刷新）
  const handleOrganizationOperation = {
    delete: async (record: API.Organization | API.OrganizationBrief) => {
      const success = await handleDelete(record);
      if (success) {
        handleTreeRefresh();
        handleTableRefresh();
      }
      return success;
    },
    toggleStatus: async (record: API.Organization | API.OrganizationBrief) => {
      const success = await handleToggleStatus(record);
      if (success) {
        handleTreeRefresh();
        handleTableRefresh();
      }
      return success;
    },
    batchDelete: async (organizations: API.OrganizationBrief[]) => {
      const success = await handleBatchDelete(organizations);
      if (success) {
        handleTreeRefresh();
        handleTableRefresh();
      }
      return success;
    },
    batchToggleStatus: async (
      organizations: API.OrganizationBrief[],
      targetStatus: "active" | "inactive"
    ) => {
      const success = await handleBatchToggleStatus(
        organizations,
        targetStatus
      );
      if (success) {
        handleTreeRefresh();
        handleTableRefresh();
      }
      return success;
    },
    syncOrganization: async (record: API.Organization | API.OrganizationBrief) => {
      try {
        const response = await syncOrganization({ id: record.id });
        if (response.code === 0) {
          message.success(
            `组织 ${record.name} 同步成功，远程组织ID: ${response.data?.remoteOrgId}`
          );
          handleTableRefresh();
        } else {
          message.error(response.msg || "同步失败");
        }
      } catch (_error) {
        message.error("同步失败");
      }
    },
    bindOrganization: (record: API.Organization | API.OrganizationBrief) => {
      setCurrentRow(record as API.Organization);
      setModalOpen("bind", true);
    },
    unbindOrganization: async (record: API.Organization | API.OrganizationBrief) => {
      Modal.confirm({
        title: `确定要解除组织 ${record.name} 的绑定吗？`,
        content: "解除绑定后将无法在 Skylark 中同步该组织，但不会删除远程组织数据",
        onOk: async () => {
          try {
            const response = await unbindOrganization({ id: record.id });
            if (response.code === 0) {
              message.success("解绑成功");
              handleTableRefresh();
            } else {
              message.error(response.msg || "解绑失败");
            }
          } catch (_error) {
            message.error("解绑失败");
          }
        },
      });
    },
  };

  // 页面额外操作按钮已移动到组织列表表格中

  return (
    <PageContainer
      header={{
        title: "组织管理",
        subTitle: "管理组织架构、成员关系和权限分配",
      }}
    >
      {/* 左右分栏布局 */}
      <Splitter layout={layout} className={styles.splitterContainer}>
        {/* 左侧组织树 */}
        <Splitter.Panel
          defaultSize="35%"
          min={380}
          max="50%"
          collapsible={{ start: true }}
          className={styles.leftPanel}
        >
          <ProCard title="组织架构">
            <OrganizationTree
              treeData={treeData}
              expandedKeys={expandedKeys}
              selectedKeys={selectedKeys}
              autoExpandParent={autoExpandParent}
              loading={treeLoading}
              onSelect={handleTreeSelect}
              onExpand={onExpand}
              onDragStart={onDragStart}
              onDragEnter={onDragEnter}
              onDrop={onDrop}
              onRefresh={handleTreeRefresh}
              onCreateOrganization={(parentId) => {
                if (parentId) {
                  const parentNode = treeData.find(
                    (node) => node.id === parentId
                  );
                  if (parentNode) {
                    setCurrentRow(parentNode as API.Organization);
                  }
                  setCreateParentId(parentId);
                } else {
                  setCreateParentId(undefined);
                }
                setModalOpen("create", true);
              }}
              onEditOrganization={handleEditClick}
              onDeleteOrganization={handleOrganizationOperation.delete}
              onViewMembers={handleMembersClick}
              onMoveOrganization={handleMoveClick}
              onExpandAll={expandAll}
              onCollapseAll={collapseAll}
            />
          </ProCard>
        </Splitter.Panel>

        {/* 右侧组织列表 */}
        <Splitter.Panel min={600} max="80%" className={styles.rightPanel}>
          <ProCard title="组织列表">
            <OrganizationTable
              actionRef={tableActionRef}
              selectionState={selectionState}
              onSelectionChange={(state) =>
                setSelectionState({
                  selectedRowKeys: state.selectedRowKeys || [],
                  selectedRows: state.selectedRows || [],
                })
              }
              onCreateOrganization={(parentId) => {
                // 如果指定了parentId，设置创建表单的初始父组织
                setCreateParentId(parentId);
                setModalOpen("create", true);
              }}
              onEditOrganization={handleEditClick}
              onDeleteOrganization={handleOrganizationOperation.delete}
              onViewMembers={handleMembersClick}
              onMoveOrganization={handleMoveClick}
              onBatchDelete={handleOrganizationOperation.batchDelete}
              onBatchToggleStatus={handleOrganizationOperation.batchToggleStatus}
              onSyncOrganization={handleOrganizationOperation.syncOrganization}
              onBindOrganization={handleOrganizationOperation.bindOrganization}
              onUnbindOrganization={handleOrganizationOperation.unbindOrganization}
              selectedOrgId={getCurrentSelectedOrgId()}
            />
          </ProCard>
        </Splitter.Panel>
      </Splitter>

      {/* 创建组织表单 */}
      <CreateOrganizationForm
        open={modalState.create}
        onOpenChange={(open) => {
          setModalOpen("create", open);
          // 关闭时清除父组织ID
          if (!open) {
            setCreateParentId(undefined);
          }
        }}
        onFinish={handleCreateSuccess}
        parentId={createParentId}
      />

      {/* 编辑组织表单 */}
      <EditOrganizationForm
        open={modalState.edit}
        onOpenChange={(open) => setModalOpen("edit", open)}
        onFinish={handleUpdateSuccess}
        currentRow={currentRow}
      />

      {/* 组织成员管理模态框 */}
      <OrganizationMembers
        open={modalState.members}
        onOpenChange={(open) => setModalOpen("members", open)}
        organization={currentRow}
        onRefresh={() => {
          handleTreeRefresh();
          handleTableRefresh();
        }}
      />

      {/* 移动组织模态框 */}
      <MoveOrganizationModal
        open={modalState.move}
        onOpenChange={(open) => setModalOpen("move", open)}
        onFinish={handleMoveSuccess}
        currentOrganization={currentRow}
      />

      {/* 绑定组织表单 */}
      <BindOrganizationForm
        open={modalState.bind}
        onOpenChange={(open) => setModalOpen("bind", open)}
        currentOrganization={currentRow}
        onSuccess={handleTableRefresh}
      />
    </PageContainer>
  );
};

export default OrganizationManagement;
