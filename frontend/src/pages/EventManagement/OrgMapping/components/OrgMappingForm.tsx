import type {
  CreateOrgMappingRequest,
  OrgMapping,
  UpdateOrgMappingRequest,
} from "@/pages/EventManagement/types";
import { getOrganizationList } from "@/services/organization";
import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormSelect,
  ProFormText,
} from "@ant-design/pro-components";
import { Spin } from "antd";
import React, { useEffect, useRef, useState } from "react";

interface OrgMappingFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (
    values: CreateOrgMappingRequest | UpdateOrgMappingRequest
  ) => Promise<boolean>;
  currentRow?: OrgMapping;
  isEditMode?: boolean;
}

const OrgMappingForm: React.FC<OrgMappingFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
  isEditMode,
}) => {
  const formRef = useRef<ProFormInstance>();
  const [organizations, setOrganizations] = useState<
    { label: string; value: string }[]
  >([]);
  const [loading, setLoading] = useState(false);

  const computedIsEdit = isEditMode ?? !!currentRow;
  const title = computedIsEdit ? "编辑组织映射" : "新建组织映射";

  // 定义初始状态常量
  const INITIAL_ORGANIZATIONS: { label: string; value: string }[] = [];

  useEffect(() => {
    if (open) {
      setLoading(true);
      getOrganizationList({ pageSize: 1000 })
        .then((orgRes) => {
          if (orgRes.code === 0 && orgRes.data?.list) {
            setOrganizations(
              orgRes.data.list.map((item) => ({
                label: `${item.name} (${item.id})`,
                value: item.id,
              }))
            );
          }
        })
        .finally(() => setLoading(false));
    }
  }, [open]);

  useEffect(() => {
    if (open && currentRow) {
      formRef.current?.setFieldsValue(currentRow);
    }
  }, [open, currentRow]);

  const handleFinish = async (values: any) => {
    try {
      const result = await onFinish(values);
      return result;
    } catch (error) {
      console.error("提交失败:", error);
      return false;
    }
  };

  return (
    <DrawerForm
      title={title}
      formRef={formRef}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
          setOrganizations(INITIAL_ORGANIZATIONS);
        }
        onOpenChange(visible);
      }}
      onFinish={handleFinish}
      preserve={false}
      drawerProps={{
        destroyOnHidden: true,
      }}
      width={600}
      layout="horizontal"
      labelCol={{ span: 6 }}
      wrapperCol={{ span: 18 }}
      submitTimeout={2000}
    >
      <Spin spinning={loading}>
        <ProFormText name="id" hidden />
        <ProFormText
          name="remoteOrgValue"
          label="远程组织值"
          rules={[{ required: true, message: "请输入远程组织字段值" }]}
          placeholder="请输入Skylark中的组织字段值，如：自贡网络维护中心地市仓库"
          tooltip="Skylark流程数据中组织字段的实际值，用于匹配和映射"
        />
        <ProFormSelect
          name="localOrgId"
          label="本地组织"
          options={organizations}
          rules={[{ required: true, message: "请选择本地组织" }]}
          placeholder="请选择要映射到的本地组织"
          showSearch
          tooltip="选择远程组织值对应的本地组织"
        />
      </Spin>
    </DrawerForm>
  );
};

export default OrgMappingForm;
