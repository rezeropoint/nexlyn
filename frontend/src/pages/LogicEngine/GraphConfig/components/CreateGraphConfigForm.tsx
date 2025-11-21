/**
 * 创建逻辑图配置表单组件
 */

import { useModel } from '@@/exports';
import {
  DrawerForm,
  ProForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { App } from 'antd';
import React, { useRef } from 'react';
import { listTags } from '@/services/lynxmanager/api';
import { DEFAULT_VERSION } from '../constants';
import IconSelector from './IconSelector';

interface CreateGraphConfigFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: any) => Promise<boolean>;
}

/**
 * 创建逻辑图配置表单
 */
const CreateGraphConfigForm: React.FC<CreateGraphConfigFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const { message } = App.useApp();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const formRef = useRef<ProFormInstance>();

  // 表单提交处理
  const handleSubmit = async (values: any) => {
    // 自动填充租户ID（从tenantInfo对象中获取）
    const tenantId = currentUser?.tenantInfo?.tenantId;
    if (!tenantId) {
      message.error('获取租户信息失败');
      return false;
    }

    // 自动填充组织ID（organizationIds数组的第一个值是主组织ID）
    const orgId = currentUser?.organizationIds?.[0];
    if (!orgId) {
      message.error('获取用户组织信息失败');
      return false;
    }

    // 构建请求数据（仅包含基本信息，节点和边在详情页编辑）
    const requestData = {
      tenantId,
      orgId, // 组织ID（用于权限控制）
      name: values.name,
      version: values.version || DEFAULT_VERSION,
      description: values.description || '',
      tagIds: values.tagIds || [],
      isEnabled: values.isEnabled !== undefined ? values.isEnabled : true,
      icon: values.icon, // 图标名称
      nodes: [], // 空数组，节点在详情页的可视化编辑器中添加
      edges: [], // 空数组，边在详情页的可视化编辑器中添加
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="新建逻辑图配置"
      width={720}
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
      layout="horizontal"
      labelCol={{ span: 3 }}
      wrapperCol={{ span: 21 }}
      initialValues={{
        version: DEFAULT_VERSION,
        isEnabled: true,
      }}
    >
      <ProFormText
        name="name"
        label="图名称"
        placeholder="请输入逻辑图名称（如 温度告警图 或 temperature_alert_graph）"
        rules={[
          { required: true, message: '请输入图名称' },
          { max: 100, message: '图名称不能超过100个字符' },
        ]}
        tooltip="支持中文、英文、数字等字符"
      />

      <ProFormText
        name="version"
        label="版本号"
        placeholder="请输入版本号（如 v1）"
        rules={[
          { required: true, message: '请输入版本号' },
          {
            pattern: /^v\d+$/,
            message: '版本号格式应为 v1, v2, v3 等',
          },
        ]}
        tooltip="版本号格式: v1, v2, v3 等"
      />

      <ProForm.Item
        name="icon"
        label="图标"
        tooltip="选择一个图标来标识此逻辑图的用途"
      >
        <IconSelector />
      </ProForm.Item>

      <ProFormTextArea
        name="description"
        label="描述"
        placeholder="请输入逻辑图描述"
        fieldProps={{ rows: 3 }}
      />

      <ProFormSelect
        name="tagIds"
        label="标签"
        mode="multiple"
        placeholder="请选择标签"
        tooltip="标签用于分类和筛选，支持多选"
        request={async () => {
          try {
            const tenantId = currentUser?.tenantInfo?.tenantId;
            if (!tenantId) {
              return [];
            }
            const { data } = await listTags({
              tenantId,
              scope: 'logic_graph',
              pageSize: 100,
            });
            return (
              data?.list?.map((tag) => ({
                label: tag.name,
                value: tag.id,
              })) || []
            );
          } catch (error) {
            console.error('加载标签列表失败:', error);
            return [];
          }
        }}
        fieldProps={{
          showSearch: true,
          optionFilterProp: 'label',
        }}
      />

      <ProFormSwitch
        name="isEnabled"
        label="启用状态"
        tooltip="是否启用此逻辑图"
        checkedChildren="启用"
        unCheckedChildren="禁用"
      />
    </DrawerForm>
  );
};

export default CreateGraphConfigForm;
