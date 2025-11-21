/**
 * Tab4: Webhooks配置（Mock数据版本）
 * 展示Webhook列表和配置表单
 */

import { PlusOutlined, InfoCircleOutlined } from '@ant-design/icons';
import {
  ModalForm,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { Alert, Badge, Button, Card, App, Popconfirm, Tag } from 'antd';
import dayjs from 'dayjs';
import React, { useState } from 'react';

interface WebhooksTabProps {
  graphId: string;
}

interface Webhook {
  id: string;
  name: string;
  url: string;
  method: string;
  trigger: string;
  headers?: string;
  bodyTemplate?: string;
  isEnabled: boolean;
  createdAt: number;
}

/**
 * Webhooks配置Tab组件
 */
const WebhooksTab: React.FC<WebhooksTabProps> = () => {
  const { message } = App.useApp();
  const [webhookModalOpen, setWebhookModalOpen] = useState(false);
  const [editingWebhook, setEditingWebhook] = useState<Webhook | null>(null);

  // Mock Webhook数据
  const [webhooks, setWebhooks] = useState<Webhook[]>([
    {
      id: '1',
      name: '执行完成通知',
      url: 'https://api.example.com/notify',
      method: 'POST',
      trigger: 'on_success',
      headers: '{"Content-Type": "application/json"}',
      bodyTemplate: '{"graphId": "{{graphId}}", "status": "{{status}}"}',
      isEnabled: true,
      createdAt: Date.now() - 86400000,
    },
    {
      id: '2',
      name: '执行失败告警',
      url: 'https://api.example.com/alert',
      method: 'POST',
      trigger: 'on_failure',
      headers: '{"Authorization": "Bearer xxx"}',
      bodyTemplate: '{"error": "{{error}}"}',
      isEnabled: true,
      createdAt: Date.now() - 172800000,
    },
  ]);

  // 触发条件选项
  const triggerOptions = [
    { label: '执行成功时', value: 'on_success' },
    { label: '执行失败时', value: 'on_failure' },
    { label: '执行完成时（无论成功失败）', value: 'on_complete' },
    { label: '执行超时时', value: 'on_timeout' },
  ];

  // 添加/编辑Webhook
  const handleSaveWebhook = async (values: any) => {
    if (editingWebhook) {
      // 编辑
      setWebhooks((prev) =>
        prev.map((item) =>
          item.id === editingWebhook.id ? { ...item, ...values } : item
        )
      );
      message.success('Webhook更新成功（Mock）');
    } else {
      // 新建
      const newWebhook: Webhook = {
        id: Date.now().toString(),
        ...values,
        createdAt: Date.now(),
      };
      setWebhooks((prev) => [...prev, newWebhook]);
      message.success('Webhook创建成功（Mock）');
    }
    setEditingWebhook(null);
    return true;
  };

  // 删除Webhook
  const handleDeleteWebhook = (record: Webhook) => {
    setWebhooks((prev) => prev.filter((item) => item.id !== record.id));
    message.success('Webhook删除成功（Mock）');
  };

  // 测试Webhook
  const handleTestWebhook = (record: Webhook) => {
    message.info(`测试Webhook: ${record.name}（功能待实现）`);
  };

  return (
    <>
      <Alert
        message="功能说明"
        description="当前为Mock数据版本，所有操作仅在前端生效，不会保存到后端。后续版本将对接真实的Webhook配置API。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        closable
        style={{ marginBottom: 16 }}
      />

      <Card
        title="Webhooks配置"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditingWebhook(null);
              setWebhookModalOpen(true);
            }}
          >
            添加Webhook
          </Button>
        }
      >
        <ProTable
          dataSource={webhooks}
          columns={[
            {
              title: '名称',
              dataIndex: 'name',
              width: 200,
            },
            {
              title: 'URL',
              dataIndex: 'url',
              ellipsis: true,
              copyable: true,
            },
            {
              title: '触发条件',
              dataIndex: 'trigger',
              width: 200,
              render: (_, record) => {
                const option = triggerOptions.find((o) => o.value === record.trigger);
                return <Tag color="blue">{option?.label}</Tag>;
              },
            },
            {
              title: '请求方法',
              dataIndex: 'method',
              width: 100,
            },
            {
              title: '状态',
              dataIndex: 'isEnabled',
              width: 100,
              render: (_, record) => (
                <Badge
                  status={record.isEnabled ? 'success' : 'default'}
                  text={record.isEnabled ? '已启用' : '已禁用'}
                />
              ),
            },
            {
              title: '创建时间',
              dataIndex: 'createdAt',
              width: 180,
              render: (_, record) => dayjs(record.createdAt).format('YYYY-MM-DD HH:mm:ss'),
            },
            {
              title: '操作',
              valueType: 'option',
              width: 150,
              render: (_, record) => [
                <a
                  key="edit"
                  onClick={() => {
                    setEditingWebhook(record);
                    setWebhookModalOpen(true);
                  }}
                >
                  编辑
                </a>,
                <a key="test" onClick={() => handleTestWebhook(record)}>
                  测试
                </a>,
                <Popconfirm
                  key="delete"
                  title="确认删除?"
                  onConfirm={() => handleDeleteWebhook(record)}
                >
                  <a>删除</a>
                </Popconfirm>,
              ],
            },
          ]}
          search={false}
          pagination={{ pageSize: 10 }}
          options={false}
        />
      </Card>

      {/* Webhook表单Modal */}
      <ModalForm
        title={editingWebhook ? '编辑Webhook' : '添加Webhook'}
        open={webhookModalOpen}
        onOpenChange={setWebhookModalOpen}
        onFinish={handleSaveWebhook}
        initialValues={editingWebhook || {}}
      >
        <ProFormText
          name="name"
          label="名称"
          rules={[{ required: true, message: '请输入名称' }]}
        />

        <ProFormText
          name="url"
          label="URL"
          rules={[
            { required: true, message: '请输入URL' },
            { type: 'url', message: '请输入有效的URL' },
          ]}
        />

        <ProFormSelect
          name="method"
          label="请求方法"
          options={[
            { label: 'GET', value: 'GET' },
            { label: 'POST', value: 'POST' },
            { label: 'PUT', value: 'PUT' },
          ]}
          rules={[{ required: true, message: '请选择请求方法' }]}
        />

        <ProFormSelect
          name="trigger"
          label="触发条件"
          options={triggerOptions}
          rules={[{ required: true, message: '请选择触发条件' }]}
        />

        <ProFormTextArea
          name="headers"
          label="请求头（JSON）"
          fieldProps={{ rows: 3 }}
          placeholder='{"Content-Type": "application/json"}'
        />

        <ProFormTextArea
          name="bodyTemplate"
          label="请求体模板"
          fieldProps={{ rows: 5 }}
          placeholder='{"graphId": "{{graphId}}", "status": "{{status}}"}'
        />

        <ProFormSwitch name="isEnabled" label="启用状态" />
      </ModalForm>
    </>
  );
};

export default WebhooksTab;
