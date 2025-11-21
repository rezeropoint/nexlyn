import type { ProFormInstance } from "@ant-design/pro-components";
import {
  DrawerForm,
  ProFormDigit,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { Divider } from "antd";
import React, { useRef } from "react";
import styles from "./CreatePlatformForm.less";

interface CreatePlatformFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: any) => Promise<boolean>;
}

/**
 * 创建平台配置表单组件
 */
const CreatePlatformForm: React.FC<CreatePlatformFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const formRef = useRef<ProFormInstance>();

  return (
    <DrawerForm
      title="新建平台配置"
      width={600}
      open={open}
      formRef={formRef}
      onOpenChange={(visible) => {
        if (!visible) {
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      onFinish={onFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="name"
        label="配置名称"
        placeholder="例如：生产环境Skylark"
        tooltip="配置的显示名称"
        rules={[{ required: true, message: "请输入配置名称" }]}
      />

      <ProFormSelect
        name="type"
        label="平台类型"
        tooltip="当前仅支持Skylark流程引擎"
        disabled
        initialValue="skylark"
        options={[
          { label: "Skylark流程引擎", value: "skylark" },
          { label: "Webhook（预留）", value: "webhook", disabled: true },
          { label: "Kafka（预留）", value: "kafka", disabled: true },
        ]}
      />

      <ProFormSwitch
        name="enabled"
        label="启用状态"
        tooltip="是否启用该平台配置"
        initialValue={true}
      />

      <ProFormTextArea
        name="description"
        label="配置描述"
        placeholder="请输入配置描述"
        fieldProps={{ rows: 3 }}
      />

      <Divider orientation="left" className={styles.divider}>
        Skylark平台配置
      </Divider>

      <ProFormText
        name="skylarkDomain"
        label="Skylark域名"
        placeholder="例如：skylark.example.com 或 localhost:8080"
        tooltip="Skylark服务的域名（不含http://或https://前缀，系统会自动添加https://）"
        rules={[
          { required: true, message: "请输入Skylark域名" },
          {
            pattern:
              /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*(:[0-9]{1,5})?$|^(\d{1,3}\.){3}\d{1,3}(:[0-9]{1,5})?$/,
            message: "请输入有效的域名或IP地址（可选端口号）",
          },
        ]}
      />

      <ProFormText
        name="skylarkAuthHeader"
        label="认证头"
        placeholder="例如：Bearer xxx"
        tooltip="Skylark API的认证Token"
        rules={[{ required: true, message: "请输入认证头" }]}
      />

      <ProFormDigit
        name="skylarkUserID"
        label="用户ID"
        placeholder="例如：123"
        tooltip="Skylark平台的用户ID"
        rules={[{ required: true, message: "请输入用户ID" }]}
        fieldProps={{ precision: 0, min: 1 }}
      />
    </DrawerForm>
  );
};

export default CreatePlatformForm;
