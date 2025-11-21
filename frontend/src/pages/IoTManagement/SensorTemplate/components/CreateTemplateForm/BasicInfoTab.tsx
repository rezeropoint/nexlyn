import type { DeviceCategoryInfo } from "@/services/iot";
import {
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React from "react";

interface BasicInfoTabProps {
  deviceCategories: DeviceCategoryInfo[];
  categoriesLoading: boolean;
  onCategoryChange: (category: string) => void;
}

/**
 * 基础信息标签页组件
 */
const BasicInfoTab: React.FC<BasicInfoTabProps> = ({
  deviceCategories,
  categoriesLoading,
  onCategoryChange,
}) => {
  return (
    <>
      <ProFormText
        name="model"
        label="设备型号"
        placeholder="请输入设备型号标识（全局唯一）"
        rules={[
          { required: true, message: "请输入设备型号" },
          {
            pattern: /^[a-zA-Z0-9_-]+$/,
            message: "只能包含字母、数字、下划线和中划线",
          },
        ]}
        tooltip="设备型号是全局唯一的标识符，只能包含字母、数字、下划线和中划线"
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
        tooltip="选择设备类别后，系统将自动加载该类别的标准字段定义"
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

export default BasicInfoTab;
