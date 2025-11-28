import { getOrganizationMembers } from "@/services/organization";
import { formatDateTime } from "@/utils/date";
import {
  CrownOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  PlusOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import {
  type ActionType,
  ModalForm,
  type ProColumns,
  ProTable,
} from "@ant-design/pro-components";
import { useAccess } from "@umijs/max";
import {
  App,
  Button,
  Card,
  Descriptions,
  message,
  Space,
  Tabs,
  Tag,
  Typography,
} from "antd";
import React, { useRef, useState } from "react";
import {
  ORGANIZATION_STATUS_COLORS,
  ORGANIZATION_STATUS_TEXT,
  RELATION_TYPE_COLORS,
  RELATION_TYPE_OPTIONS,
  RELATION_TYPE_TEXT,
  TABLE_CONFIG,
} from "../constants";
import { useOrganizationMembers } from "../hooks/useOrganizationMembers";
import AddMemberForm from "./AddMemberForm";
import EditMemberForm from "./EditMemberForm";
import styles from "./OrganizationMembers.less";

const { Title, Text } = Typography;
const { TabPane } = Tabs;

interface OrganizationMembersProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organization?: API.Organization | API.OrganizationBrief;
  onRefresh?: () => void;
}

/**
 * 组织成员管理组件
 */
const OrganizationMembers: React.FC<OrganizationMembersProps> = ({
  open,
  onOpenChange,
  organization,
  onRefresh,
}) => {
  const { modal } = App.useApp();
  const access = useAccess();
  const memberActionRef = useRef<ActionType>();
  const [activeTab, setActiveTab] = useState<string>("members");

  // 使用组织成员管理Hook
  const {
    currentMember,
    memberModalState,
    memberSelectionState,
    setMemberModalOpen,
    setMemberSelectionState,
    handleAddMember,
    handleRemoveMember,
    handleUpdateMember,
    handleBatchRemoveMembers,
    handleBatchUpdateMemberType,
    handleEditMemberClick,
    handleSetAsManager,
  } = useOrganizationMembers(organization?.id || "");

  if (!organization) return null;

  // 单个移除成员确认
  const handleRemoveClick = (member: API.UserOrgRelation) => {
    modal.confirm({
      title: "确认移除成员",
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>
            确定要将 <strong>{member.name}</strong> 从组织{" "}
            <strong>{organization.name}</strong> 中移除吗？
          </p>
          <p className={styles.hintText}>
            移除后该用户将无法访问此组织的相关资源。
          </p>
        </div>
      ),
      okText: "确认移除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        const success = await handleRemoveMember(member);
        if (success) {
          memberActionRef.current?.reload();
          onRefresh?.();
        }
      },
    });
  };

  // 批量移除成员确认
  const handleBatchRemoveClick = () => {
    if (memberSelectionState.selectedRows.length === 0) {
      message.warning("请先选择要移除的成员");
      return;
    }

    modal.confirm({
      title: "确认批量移除",
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>
            确定要移除选中的{" "}
            <strong>{memberSelectionState.selectedRows.length}</strong>{" "}
            个成员吗？
          </p>
          <p className={styles.hintText}>
            移除后这些用户将无法访问此组织的相关资源。
          </p>
        </div>
      ),
      okText: "确认移除",
      okType: "danger",
      cancelText: "取消",
      onOk: async () => {
        const success = await handleBatchRemoveMembers(
          memberSelectionState.selectedRows
        );
        if (success) {
          memberActionRef.current?.reload();
          onRefresh?.();
        }
      },
    });
  };

  // 设置为负责人确认
  const handleSetManagerClick = (member: API.UserOrgRelation) => {
    modal.confirm({
      title: "确认设置负责人",
      content: (
        <div>
          <p>
            确定要将 <strong>{member.name}</strong> 设置为组织{" "}
            <strong>{organization.name}</strong> 的负责人吗？
          </p>
          <p className={styles.hintText}>设置后原负责人将变为普通成员。</p>
        </div>
      ),
      okText: "确认设置",
      cancelText: "取消",
      onOk: async () => {
        const success = await handleSetAsManager(member);
        if (success) {
          memberActionRef.current?.reload();
          onRefresh?.();
        }
      },
    });
  };

  // 成员表格列定义
  const memberColumns: ProColumns<API.UserOrgRelation>[] = [
    {
      title: "姓名",
      dataIndex: "name",
      width: 100,
      fixed: "left",
      render: (name, record) => (
        <Space>
          <span>{name}</span>
          {record.relationType === "manager" && (
            <CrownOutlined className={styles.managerIcon} title="组织负责人" />
          )}
        </Space>
      ),
    },
    {
      title: "用户名",
      dataIndex: "userName",
      width: 120,
      copyable: true,
    },
    {
      title: "邮箱",
      dataIndex: "email",
      width: 180,
      copyable: true,
      ellipsis: true,
    },
    {
      title: "职位",
      dataIndex: "positionTitle",
      width: 120,
      render: (positionTitle) => positionTitle || "-",
    },
    {
      title: "关系类型",
      dataIndex: "relationType",
      width: 100,
      valueType: "select",
      valueEnum: RELATION_TYPE_OPTIONS.reduce((acc, item) => {
        acc[item.value] = { text: item.label };
        return acc;
      }, {} as Record<string, { text: string }>),
      render: (_, record) => (
        <Tag
          color={
            RELATION_TYPE_COLORS[
              record.relationType as keyof typeof RELATION_TYPE_COLORS
            ]
          }
        >
          {
            RELATION_TYPE_TEXT[
              record.relationType as keyof typeof RELATION_TYPE_TEXT
            ]
          }
        </Tag>
      ),
    },
    {
      title: "主组织",
      dataIndex: "isPrimary",
      width: 80,
      valueType: "select",
      valueEnum: {
        true: { text: "是" },
        false: { text: "否" },
      },
      render: (isPrimary) => (
        <Tag color={isPrimary ? "success" : "default"}>
          {isPrimary ? "是" : "否"}
        </Tag>
      ),
    },
    {
      title: "加入时间",
      dataIndex: "joinedAt",
      width: 160,
      hideInSearch: true,
      render: (_, record) => formatDateTime(record.joinedAt),
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      width: 180,
      fixed: "right",
      render: (_, record) => (
        <Space size="small">
          {access.canManageOrganizationMembers() && (
            <>
              <Button
                type="link"
                size="small"
                icon={<EditOutlined />}
                onClick={() => handleEditMemberClick(record)}
              >
                编辑
              </Button>

              {record.relationType !== "manager" && (
                <Button
                  type="link"
                  size="small"
                  icon={<CrownOutlined />}
                  onClick={() => handleSetManagerClick(record)}
                >
                  设为负责人
                </Button>
              )}

              <Button
                type="link"
                size="small"
                danger
                icon={<DeleteOutlined />}
                onClick={() => handleRemoveClick(record)}
              >
                移除
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  // 成员表格工具栏
  const memberToolBarRender = () => {
    const hasSelected = memberSelectionState.selectedRowKeys.length > 0;

    return [
      // 多选操作按钮
      hasSelected && access.canManageOrganizationMembers() && (
        <Space key="batch-actions">
          <Button
            onClick={() =>
              handleBatchUpdateMemberType(
                memberSelectionState.selectedRows,
                "member"
              )
            }
            disabled={
              !memberSelectionState.selectedRows.some(
                (m) => m.relationType !== "member"
              )
            }
          >
            设为普通成员
          </Button>
          <Button danger onClick={handleBatchRemoveClick}>
            批量移除
          </Button>
        </Space>
      ),

      // 添加成员按钮
      access.canManageOrganizationMembers() && (
        <Button
          key="add"
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setMemberModalOpen("add", true)}
        >
          添加成员
        </Button>
      ),
    ].filter(Boolean);
  };

  // 处理添加成员成功
  const handleAddMemberSuccess = async (values: any): Promise<boolean> => {
    const success = await handleAddMember(values);
    if (success) {
      memberActionRef.current?.reload();
      onRefresh?.();
    }
    return success;
  };

  // 处理编辑成员成功
  const handleEditMemberSuccess = async (values: any): Promise<boolean> => {
    const success = await handleUpdateMember(values);
    if (success) {
      memberActionRef.current?.reload();
      onRefresh?.();
    }
    return success;
  };

  return (
    <>
      <ModalForm
        title={
          <Space>
            <TeamOutlined />
            <span>{organization.name} - 成员管理</span>
          </Space>
        }
        open={open}
        onOpenChange={onOpenChange}
        submitter={false}
        modalProps={{
          width: 1200,
          destroyOnHidden: true,
        }}
      >
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="成员列表" key="members">
            <ProTable<API.UserOrgRelation>
              actionRef={memberActionRef}
              columns={memberColumns}
              request={async (params) => {
                try {
                  const response = await getOrganizationMembers({
                    id: organization.id,
                    current: params.current || 1,
                    pageSize: params.pageSize || TABLE_CONFIG.PAGE_SIZE,
                    keyword: params.keyword,
                    relationType: params.relationType,
                    isPrimary: params.isPrimary,
                  });

                  if (response.code === 0) {
                    return {
                      data: response.data?.list || [],
                      success: true,
                      total: response.total || 0,
                    };
                  } else {
                    message.error(response.msg || "获取成员列表失败");
                    return {
                      data: [],
                      success: false,
                      total: 0,
                    };
                  }
                } catch (error) {
                  console.error("获取成员列表失败:", error);
                  message.error("获取成员列表失败");
                  return {
                    data: [],
                    success: false,
                    total: 0,
                  };
                }
              }}
              rowKey="id"
              search={{
                labelWidth: "auto",
                defaultCollapsed: false,
              }}
              pagination={{
                defaultPageSize: TABLE_CONFIG.PAGE_SIZE,
                showSizeChanger: true,
                pageSizeOptions: [...TABLE_CONFIG.PAGE_SIZE_OPTIONS],
              }}
              scroll={{ x: 1000 }}
              toolBarRender={memberToolBarRender}
              rowSelection={
                access.canManageOrganizationMembers()
                  ? {
                      selectedRowKeys: memberSelectionState.selectedRowKeys,
                      onChange: (keys, rows) => {
                        setMemberSelectionState({
                          selectedRowKeys: keys,
                          selectedRows: rows,
                        });
                      },
                    }
                  : false
              }
              tableAlertRender={({ selectedRowKeys }) => (
                <Space size={24}>
                  <span>
                    已选择 <strong>{selectedRowKeys.length}</strong> 项
                  </span>
                </Space>
              )}
            />
          </TabPane>

          <TabPane tab="组织信息" key="info">
            <Card>
              <Descriptions title="基本信息" column={2}>
                <Descriptions.Item label="组织名称">
                  {organization.name}
                </Descriptions.Item>
                <Descriptions.Item label="组织代码">
                  {organization.code}
                </Descriptions.Item>
                <Descriptions.Item label="组织类型">
                  {organization.type}
                </Descriptions.Item>
                <Descriptions.Item label="层级">
                  第{organization.level}层
                </Descriptions.Item>
                <Descriptions.Item label="路径">
                  {organization.path}
                </Descriptions.Item>
                <Descriptions.Item label="负责人">
                  {organization.managerName || "未设置"}
                </Descriptions.Item>
                <Descriptions.Item label="成员数量">
                  {organization.memberCount}人
                </Descriptions.Item>
                <Descriptions.Item label="状态">
                  <Tag
                    color={
                      ORGANIZATION_STATUS_COLORS[
                        organization.status as keyof typeof ORGANIZATION_STATUS_COLORS
                      ]
                    }
                  >
                    {
                      ORGANIZATION_STATUS_TEXT[
                        organization.status as keyof typeof ORGANIZATION_STATUS_TEXT
                      ]
                    }
                  </Tag>
                </Descriptions.Item>
              </Descriptions>

              {"description" in organization && organization.description && (
                <div className={styles.descriptionSection}>
                  <Title level={5}>组织描述</Title>
                  <Text>{organization.description}</Text>
                </div>
              )}
            </Card>
          </TabPane>
        </Tabs>
      </ModalForm>

      {/* 添加成员表单 */}
      <AddMemberForm
        open={memberModalState.add}
        onOpenChange={(open) => setMemberModalOpen("add", open)}
        onFinish={handleAddMemberSuccess}
        organization={organization}
      />

      {/* 编辑成员表单 */}
      <EditMemberForm
        open={memberModalState.edit}
        onOpenChange={(open) => setMemberModalOpen("edit", open)}
        onFinish={handleEditMemberSuccess}
        currentMember={currentMember}
        organization={organization}
      />
    </>
  );
};

export default OrganizationMembers;
