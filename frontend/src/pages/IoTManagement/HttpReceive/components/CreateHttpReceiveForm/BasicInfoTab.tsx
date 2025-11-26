import {
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React from "react";

/**
 * 基础信息标签页组件
 *
 * HTTP 接收功能处理任意格式的 JSON 数据，不需要设备类别和型号字段
 */
const BasicInfoTab: React.FC = () => {
  return (
    <>
      <ProFormText
        name="name"
        label="配置名称"
        placeholder="请输入配置名称"
        rules={[{ required: true, message: "请输入配置名称" }]}
        tooltip="用于标识此 HTTP 接收配置的名称"
      />

      <ProFormSwitch
        name="enabled"
        label="启用状态"
        fieldProps={{
          checkedChildren: "已启用",
          unCheckedChildren: "已禁用",
        }}
        initialValue={true}
        tooltip="禁用后，此配置对应的数据接收端点将不再处理请求"
      />

      <ProFormTextArea
        name="description"
        label="配置描述"
        placeholder="请输入配置描述（可选）"
        fieldProps={{
          rows: 3,
        }}
      />
    </>
  );
};

export default BasicInfoTab;
