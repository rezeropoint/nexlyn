/**
 * 编辑标签表单组件
 */

import type { LynxTag } from '@/services/lynxmanager';
import { useModel } from '@@/exports';
import {
  DrawerForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { App } from 'antd';
import React, { useRef } from 'react';
import { TAG_SCOPE_OPTIONS } from '../constants';

interface EditTagFormProps {
  open: boolean;
  data?: LynxTag;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: any) => Promise<boolean>;
}

/**
 * 编辑标签表单
 */
const EditTagForm: React.FC<EditTagFormProps> = ({
  open,
  data,
  onOpenChange,
  onFinish,
}) => {
  const { message } = App.useApp();
  const { initialState } = useModel('@@initialState');
  const _currentUser = initialState?.currentUser;
  const formRef = useRef<ProFormInstance>();

  // 表单提交处理
  const handleSubmit = async (values: any) => {
    if (!data) {
      message.error('标签信息不存在');
      return false;
    }

    // 构建请求数据
    const requestData = {
      name: values.name,
      scope: values.scope,
      description: values.description,
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="编辑标签"
      width={600}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置表单
          formRef.current?.resetFields();
        }
        onOpenChange(visible);
      }}
      formRef={formRef}
      onFinish={handleSubmit}
      autoFocusFirstInput
      drawerProps={{
        destroyOnHidden: true,
      }}
      layout="vertical"
      initialValues={{
        name: data?.name,
        scope: data?.scope,
        description: data?.description,
      }}
    >
      <ProFormText
        name="name"
        label="标签名称"
        placeholder="请输入标签名称"
        rules={[
          { required: true, message: '请输入标签名称' },
          { max: 100, message: '标签名称不超过100个字符' },
        ]}
      />

      <ProFormSelect
        name="scope"
        label="作用域"
        placeholder="请选择标签作用域"
        options={Object.values(TAG_SCOPE_OPTIONS)}
        rules={[{ required: true, message: '请选择作用域' }]}
        disabled
        tooltip="作用域在创建后不能修改"
      />

      <ProFormTextArea
        name="description"
        label="描述"
        placeholder="请输入标签描述（可选）"
        fieldProps={{
          rows: 4,
          maxLength: 500,
        }}
      />
    </DrawerForm>
  );
};

export default EditTagForm;
