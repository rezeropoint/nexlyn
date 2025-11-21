import type { DeviceCategoryInfo } from "@/services/iot";
import {
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React from "react";

interface EditBasicInfoTabProps {
  deviceCategories: DeviceCategoryInfo[];
  categoriesLoading: boolean;
  onCategoryChange: (category: string) => void;
}

/**
 * 编辑基础信息标签页组件
 * 与创建表单的区别：设备型号字段禁用
 */
const EditBasicInfoTab: React.FC<EditBasicInfoTabProps> = ({
  deviceCategories,
  categoriesLoading,
  onCategoryChange,
}) => {
  return (
    <>
      <ProFormText
        name="model"
        label="设备型号"
        disabled
        tooltip="设备型号不可修改"
      />

      <ProFormText
        name="name"
        label="模板名称"
        placeholder="请输入模板显示名称"
        rules={[{ required: true, message: "请输入模板名称" }]}
      />

      <ProFormSelect
        name="category"
        label="设备类别"
        placeholder="请选择设备类别"
        disabled
        options={deviceCategories.map((cat) => ({
          label: `${cat.name} (${cat.code})`,
          value: cat.code,
          description: cat.description,
        }))}
        fieldProps={{
          onChange: onCategoryChange,
          showSearch: true,
          optionFilterProp: "label",
          loading: categoriesLoading,
        }}
        tooltip="设备类别是模板核心标识，创建后不可修改"
      />

      <ProFormText
        name="manufacturer"
        label="设备厂商"
        placeholder="请输入设备厂商名称"
      />

      <ProFormText
        name="version"
        label="模板版本"
        placeholder="请输入模板版本号"
      />

      <ProFormTextArea
        name="description"
        label="模板描述"
        placeholder="请输入模板描述信息"
        fieldProps={{
          rows: 3,
        }}
      />

      <ProFormSwitch
        name="enabled"
        label="启用状态"
        fieldProps={{
          checkedChildren: "已启用",
          unCheckedChildren: "已禁用",
        }}
      />
    </>
  );
};

export default EditBasicInfoTab;
