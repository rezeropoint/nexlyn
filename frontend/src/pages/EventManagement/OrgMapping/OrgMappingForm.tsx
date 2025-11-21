import { createOrgMapping, updateOrgMapping } from "@/services/eventhandler";
import { getOrganizationList } from "@/services/organization";
import {
  ModalForm,
  ProFormText,
  ProFormTreeSelect,
} from "@ant-design/pro-components";
import { Form, message } from "antd";
import React, { useEffect, useState } from "react";
import { MESSAGE } from "../constants";
import type { OrgMapping } from "../types";

interface OrgMappingFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentMapping?: OrgMapping;
  onSuccess: () => void;
}

interface OrganizationNode {
  id: string;
  name: string;
  parentId: string | null;
  children?: OrganizationNode[];
}

/**
 * 组织映射表单
 */
const OrgMappingForm: React.FC<OrgMappingFormProps> = ({
  open,
  onOpenChange,
  currentMapping,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const [orgTreeData, setOrgTreeData] = useState<any[]>([]);
  const isEdit = !!currentMapping;

  // 加载组织树数据
  const loadOrganizations = async () => {
    try {
      const response = await getOrganizationList({ pageSize: 1000 });
      if (response.code === 0 && response.data) {
        const orgList = response.data.list || [];
        const treeData = buildTreeData(orgList);
        setOrgTreeData(treeData);
      }
    } catch (error) {
      console.error("获取组织列表失败:", error);
      message.error("获取组织列表失败");
    }
  };

  // 构建树形数据
  const buildTreeData = (organizations: OrganizationNode[]): any[] => {
    const map: { [key: string]: any } = {};
    const roots: any[] = [];

    // 创建节点映射
    organizations.forEach((org) => {
      map[org.id] = {
        title: org.name,
        value: org.id,
        key: org.id,
        children: [],
      };
    });

    // 构建树形结构
    organizations.forEach((org) => {
      if (org.parentId && map[org.parentId]) {
        map[org.parentId].children.push(map[org.id]);
      } else {
        roots.push(map[org.id]);
      }
    });

    return roots;
  };

  useEffect(() => {
    if (open) {
      loadOrganizations();

      if (currentMapping) {
        // 编辑模式：设置表单值
        form.setFieldsValue({
          remoteOrgValue: currentMapping.remoteOrgValue,
          localOrgId: currentMapping.localOrgId,
        });
      } else {
        // 新建模式：重置表单
        form.resetFields();
      }
    }
  }, [open, currentMapping]);

  // 提交表单
  const handleSubmit = async (values: any) => {
    try {
      const response = isEdit
        ? await updateOrgMapping(currentMapping.id, values)
        : await createOrgMapping(values);

      if (response.code === 0) {
        message.success(
          isEdit ? MESSAGE.UPDATE_SUCCESS : MESSAGE.CREATE_SUCCESS
        );
        onSuccess();
        return true;
      } else {
        message.error(
          response.msg ||
            (isEdit ? MESSAGE.UPDATE_FAILED : MESSAGE.CREATE_FAILED)
        );
        return false;
      }
    } catch (error) {
      console.error("保存组织映射失败:", error);
      message.error(isEdit ? MESSAGE.UPDATE_FAILED : MESSAGE.CREATE_FAILED);
      return false;
    }
  };

  return (
    <ModalForm
      title={isEdit ? "编辑组织映射" : "新建组织映射"}
      open={open}
      onOpenChange={onOpenChange}
      form={form}
      onFinish={handleSubmit}
      modalProps={{
        width: 600,
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="remoteOrgValue"
        label="远程组织值"
        placeholder="请输入远程系统的组织字段值"
        rules={[
          { required: true, message: "请输入远程组织值" },
          { max: 200, message: "远程组织值最多200个字符" },
        ]}
        tooltip="Skylark系统中的组织字段值，例如：自贡网络维护中心地市仓库"
      />

      <ProFormTreeSelect
        name="localOrgId"
        label="本地组织"
        placeholder="请选择本地组织"
        rules={[{ required: true, message: "请选择本地组织" }]}
        request={async () => orgTreeData}
        fieldProps={{
          showSearch: true,
          treeNodeFilterProp: "title",
          styles: { popup: { root: { maxHeight: 400, overflow: "auto" } } },
        }}
        tooltip="选择对应的本地组织，用于权限控制"
      />
    </ModalForm>
  );
};

export default OrgMappingForm;
