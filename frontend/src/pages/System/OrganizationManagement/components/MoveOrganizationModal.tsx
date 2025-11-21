import { getOrganizationTree } from "@/services/organization";
import {
  ApartmentOutlined,
  ExclamationCircleOutlined,
} from "@ant-design/icons";
import {
  Alert,
  Card,
  Descriptions,
  Form,
  InputNumber,
  Modal,
  Space,
  Tag,
  TreeSelect,
} from "antd";
import React, { useEffect, useState } from "react";
import {
  ORGANIZATION_STATUS_COLORS,
  ORGANIZATION_STATUS_TEXT,
} from "../constants";
import type { TreeNodeData } from "../types";
import styles from "./MoveOrganizationModal.less";

interface MoveOrganizationModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: {
    newParentId: string;
    newSortOrder?: number;
  }) => Promise<boolean>;
  currentOrganization?: API.Organization | API.OrganizationBrief;
}

/**
 * 移动组织模态框组件
 */
const MoveOrganizationModal: React.FC<MoveOrganizationModalProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentOrganization,
}) => {
  const [form] = Form.useForm();
  const [treeData, setTreeData] = useState<TreeNodeData[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [confirmLoading, setConfirmLoading] = useState<boolean>(false);
  const [selectedParent, setSelectedParent] = useState<TreeNodeData | null>(
    null
  );

  // 将组织树数据转换为TreeSelect需要的格式
  const convertToTreeSelectData = (
    organizations: API.OrganizationTree[]
  ): TreeNodeData[] => {
    return organizations.map((org) => ({
      ...org,
      key: org.id,
      value: org.id,
      title: `${org.name} (${org.code})`,
      // 禁用当前组织及其子组织（防止循环引用）
      disabled: isDescendantOrSelf(org, currentOrganization?.id || ""),
      children: org.children ? convertToTreeSelectData(org.children) : [],
    }));
  };

  // 检查是否为当前组织的子孙组织或自身
  const isDescendantOrSelf = (
    org: API.OrganizationTree,
    currentOrgId: string
  ): boolean => {
    if (org.id === currentOrgId) return true;
    if (!org.children) return false;
    return org.children.some((child) =>
      isDescendantOrSelf(child, currentOrgId)
    );
  };

  // 根据ID查找组织节点
  const findOrgById = (
    nodes: TreeNodeData[],
    id: string
  ): TreeNodeData | null => {
    for (const node of nodes) {
      if (node.id === id) return node;
      if (node.children && node.children.length > 0) {
        const found = findOrgById(node.children, id);
        if (found) return found;
      }
    }
    return null;
  };

  // 预测移动后的新路径
  const predictNewPath = (
    parentPath: string | undefined,
    currentOrgId: string
  ): string => {
    if (!parentPath) {
      return `/root/${currentOrgId}/`;
    }
    return `${parentPath}${currentOrgId}/`;
  };

  // 加载组织树数据
  const loadTreeData = async () => {
    setLoading(true);
    try {
      const response = await getOrganizationTree({});
      if (response.code === 0 && response.data) {
        const convertedData = convertToTreeSelectData(response.data);
        setTreeData(convertedData);
      }
    } catch (error) {
      console.error("获取组织树失败:", error);
    } finally {
      setLoading(false);
    }
  };

  // 处理父组织选择
  const handleParentChange = (value: string) => {
    const parent = value ? findOrgById(treeData, value) : null;
    setSelectedParent(parent);

    // 自动设置排序值为当前父组织下子组织数量+1
    if (parent?.children) {
      form.setFieldsValue({ newSortOrder: parent.children.length });
    } else {
      form.setFieldsValue({ newSortOrder: 0 });
    }
  };

  // 处理确认移动
  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      setConfirmLoading(true);

      const success = await onFinish({
        newParentId: values.newParentId || "", // 空字符串表示移动到根级
        newSortOrder: values.newSortOrder,
      });

      if (success) {
        form.resetFields();
        setSelectedParent(null);
        onOpenChange(false);
      }
    } catch (error) {
      console.error("移动组织失败:", error);
    } finally {
      setConfirmLoading(false);
    }
  };

  // 处理取消
  const handleCancel = () => {
    form.resetFields();
    setSelectedParent(null);
    onOpenChange(false);
  };

  // 组件打开时加载数据
  useEffect(() => {
    if (open) {
      loadTreeData();
    }
  }, [open]);

  if (!currentOrganization) return null;

  const newPath = predictNewPath(selectedParent?.path, currentOrganization.id);
  const newLevel = selectedParent ? selectedParent.level + 1 : 1;

  return (
    <Modal
      title={
        <Space>
          <ApartmentOutlined />
          <span>移动组织</span>
        </Space>
      }
      open={open}
      onOk={handleOk}
      onCancel={handleCancel}
      confirmLoading={confirmLoading}
      width={700}
      okText="确认移动"
      cancelText="取消"
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          newSortOrder: 0,
        }}
      >
        {/* 当前组织信息 */}
        <Card title="当前组织信息" size="small" style={{ marginBottom: 16 }}>
          <Descriptions column={2}>
            <Descriptions.Item label="组织名称">
              {currentOrganization.name}
            </Descriptions.Item>
            <Descriptions.Item label="组织代码">
              {currentOrganization.code}
            </Descriptions.Item>
            <Descriptions.Item label="当前路径">
              {currentOrganization.path}
            </Descriptions.Item>
            <Descriptions.Item label="当前层级">
              第{currentOrganization.level}层
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag
                color={
                  ORGANIZATION_STATUS_COLORS[
                    currentOrganization.status as keyof typeof ORGANIZATION_STATUS_COLORS
                  ]
                }
              >
                {
                  ORGANIZATION_STATUS_TEXT[
                    currentOrganization.status as keyof typeof ORGANIZATION_STATUS_TEXT
                  ]
                }
              </Tag>
            </Descriptions.Item>
          </Descriptions>
        </Card>

        {/* 移动设置 */}
        <Card title="移动设置" size="small">
          <Form.Item
            name="newParentId"
            label="新的父组织"
            tooltip="选择要移动到的父组织，不选择则移动到根级别"
          >
            <TreeSelect
              placeholder="请选择新的父组织（不选择则移动到根级）"
              allowClear
              showSearch
              treeDefaultExpandAll={false}
              treeData={treeData}
              loading={loading}
              onChange={handleParentChange}
              treeNodeFilterProp="title"
              styles={{ popup: { root: { maxHeight: 400, overflow: "auto" } } }}
            />
          </Form.Item>

          <Form.Item
            name="newSortOrder"
            label="排序值"
            tooltip="在同级组织中的排序位置，数值越小排序越靠前"
          >
            <InputNumber
              min={0}
              max={999999}
              style={{ width: "100%" }}
              placeholder="请输入排序值"
            />
          </Form.Item>

          {/* 移动预览 */}
          {(selectedParent || form.getFieldValue("newParentId") === "") && (
            <Alert
              message="移动预览"
              description={
                <div>
                  <p>
                    <strong>新的父组织：</strong>
                    {selectedParent
                      ? `${selectedParent.name} (${selectedParent.code})`
                      : "根级别"}
                  </p>
                  <p>
                    <strong>新的路径：</strong>
                    {newPath}
                  </p>
                  <p>
                    <strong>新的层级：</strong>第{newLevel}层
                  </p>
                  {newLevel > 10 && (
                    <p className={styles.warningText}>
                      <ExclamationCircleOutlined />{" "}
                      警告：移动后层级将超过10层限制！
                    </p>
                  )}
                </div>
              }
              type="info"
              className={styles.previewAlert}
            />
          )}

          {/* 注意事项 */}
          <Alert
            message="注意事项"
            description={
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                <li>移动组织会同时移动其所有子组织</li>
                <li>移动后所有子组织的路径和层级都会自动更新</li>
                <li>不能将组织移动到自己或自己的子组织下</li>
                <li>组织层级不能超过10层</li>
                <li>移动操作会影响相关权限和关联关系</li>
              </ul>
            }
            type="warning"
            style={{ marginTop: 16 }}
          />
        </Card>
      </Form>
    </Modal>
  );
};

export default MoveOrganizationModal;
