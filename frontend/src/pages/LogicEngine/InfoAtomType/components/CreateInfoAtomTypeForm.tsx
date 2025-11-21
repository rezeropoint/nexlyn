/**
 * 创建信息原子类型表单组件
 */

import { useModel } from '@@/exports';
import {
  DrawerForm,
  ProForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
} from '@ant-design/pro-components';
import type { FormInstance } from 'antd';
import { Alert, App } from 'antd';
import React, { useRef, useState } from 'react';
import { listTags } from '@/services/lynxmanager/api';
import { DEFAULT_VERSION } from '../constants';
import FieldListTable, { type FieldItem } from './FieldListTable';
import styles from './InfoAtomTypeForm.less';

interface CreateInfoAtomTypeFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: any) => Promise<boolean>;
}

/**
 * 创建信息原子类型表单
 */
const CreateInfoAtomTypeForm: React.FC<CreateInfoAtomTypeFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const { message } = App.useApp();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const formRef = useRef<ProFormInstance>();
  const [fields, setFields] = useState<FieldItem[]>([
    { fieldKey: '', fieldPath: '', fieldType: 'string' },
  ]);

  // 表单提交处理
  const handleSubmit = async (values: any) => {
    // 自动填充租户ID（从tenantInfo对象中获取）
    const tenantId = currentUser?.tenantInfo?.tenantId;
    if (!tenantId) {
      message.error('获取租户信息失败');
      return false;
    }

    // 验证字段列表
    if (fields.length === 0) {
      message.error('请至少添加一个字段');
      return false;
    }

    // 验证每个字段
    for (let i = 0; i < fields.length; i++) {
      const field = fields[i];
      if (!field.fieldKey) {
        message.error(`第 ${i + 1} 个字段的字段键不能为空`);
        return false;
      }
      if (!/^[a-zA-Z0-9_-]+$/.test(field.fieldKey)) {
        message.error(`第 ${i + 1} 个字段的字段键只能包含字母、数字、下划线和连字符`);
        return false;
      }
      if (!field.fieldPath) {
        message.error(`第 ${i + 1} 个字段的字段路径不能为空`);
        return false;
      }
    }

    // 构建请求数据
    const requestData = {
      tenantId,
      name: values.name,
      version: values.version || DEFAULT_VERSION,
      tagIds: values.tagIds || [],
      dataFormat: {
        dataPlural: values.dataPlural || false,
        fieldStart: values.fieldStart || '',
        fields,
      },
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="新建信息原子类型"
      width={800}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置表单和字段列表
          formRef.current?.resetFields();
          setFields([{ fieldKey: '', fieldPath: '', fieldType: 'string' }]);
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
      labelCol={{ span: 4 }}
      wrapperCol={{ span: 20 }}
      initialValues={{
        version: DEFAULT_VERSION,
        dataPlural: false,
      }}
    >
      <ProFormText
        name="name"
        label="类型名称"
        placeholder="请输入信息原子类型名称（如 sensor.temperature 或 传感器.温度）"
        rules={[
          { required: true, message: '请输入类型名称' },
          { max: 255, message: '类型名称不能超过255个字符' },
        ]}
        tooltip="类型名称用于标识信息原子的数据结构，支持中英文"
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
              scope: 'info_atom',
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
        name="dataPlural"
        label="数据为数组"
        tooltip="是否为数组类型的数据"
        checkedChildren="是"
        unCheckedChildren="否"
      />

      <ProForm.Item noStyle shouldUpdate>
        {(form: FormInstance) => {
          const dataPlural = form.getFieldValue('dataPlural');
          if (!dataPlural) return null;

          return (
            <ProFormText
              name="fieldStart"
              label="数组起始路径"
              placeholder="请输入数组数据的起始字段路径（JSONPath）"
              tooltip="当数据为数组时，指定数组数据在JSON中的路径"
            />
          );
        }}
      </ProForm.Item>

      <Alert
        message="字段配置"
        description="请配置信息原子的字段列表。字段键用于标识字段，字段路径支持JSONPath表达式。"
        type="info"
        className={styles.fieldConfigAlert}
      />

      <FieldListTable fields={fields} onChange={setFields} />
    </DrawerForm>
  );
};

export default CreateInfoAtomTypeForm;
