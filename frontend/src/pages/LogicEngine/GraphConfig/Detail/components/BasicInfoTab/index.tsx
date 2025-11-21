/**
 * Tab1: 基本信息页面
 * 使用Descriptions展示元数据，支持编辑
 */

import { EditOutlined } from '@ant-design/icons';
import { ModalForm, ProForm, type ProFormInstance, ProFormSwitch, ProFormSelect, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import type { GraphConfigDetail } from '@/services/lynxmanager/types';
import { listTags } from '@/services/lynxmanager/api';
import { getUserById } from '@/services/user';
import { Badge, Button, Card, Descriptions, Empty, Space, Tag, Typography, App } from 'antd';
import { useModel } from '@@/exports';
import dayjs from 'dayjs';
import React, { useState, useEffect, useRef } from 'react';
import IconSelector from '../../../components/IconSelector';
import { getIconComponent, getIconColor } from '../../../iconConfig';

const { Text } = Typography;

interface BasicInfoTabProps {
  data: GraphConfigDetail | null;
  loading: boolean;
  onUpdate: (values: Partial<GraphConfigDetail>) => Promise<boolean>;
}

/**
 * 基本信息Tab组件
 */
const BasicInfoTab: React.FC<BasicInfoTabProps> = ({ data, loading, onUpdate }) => {
  const { message } = App.useApp();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const formRef = useRef<ProFormInstance>();

  const [editModalOpen, setEditModalOpen] = useState(false);
  const [createdByName, setCreatedByName] = useState<string>('');
  const [updatedByName, setUpdatedByName] = useState<string>('');
  const [loadingUserNames, setLoadingUserNames] = useState(false);
  // 标签选项列表
  const [tagOptions, setTagOptions] = useState<Array<{ label: string; value: string }>>([]);

  // 获取用户名
  useEffect(() => {
    if (!data) return;

    const fetchUserNames = async () => {
      setLoadingUserNames(true);
      try {
        // 并行获取创建人和更新人的姓名
        const promises: Promise<any>[] = [];
        if (data.createdBy) {
          promises.push(getUserById({ id: data.createdBy }));
        }
        if (data.updatedBy && data.updatedBy !== data.createdBy) {
          promises.push(getUserById({ id: data.updatedBy }));
        }

        const results = await Promise.all(promises);

        const creatorIndex = 0;
        if (data.createdBy && results[creatorIndex]?.code === 0 && results[creatorIndex]?.data) {
          setCreatedByName(results[creatorIndex].data.name || data.createdBy);
        }

        if (data.updatedBy) {
          if (data.updatedBy === data.createdBy) {
            setUpdatedByName(createdByName || results[0]?.data?.name || data.updatedBy);
          } else if (results[1]?.code === 0 && results[1]?.data) {
            setUpdatedByName(results[1].data.name || data.updatedBy);
          }
        }
      } catch (error) {
        console.error('获取用户名失败:', error);
        // 失败时显示用户ID
        if (data.createdBy) setCreatedByName(data.createdBy);
        if (data.updatedBy) setUpdatedByName(data.updatedBy);
      } finally {
        setLoadingUserNames(false);
      }
    };

    fetchUserNames();
  }, [data?.createdBy, data?.updatedBy]);

  // 加载标签列表
  useEffect(() => {
    if (editModalOpen) {
      const tenantId = currentUser?.tenantInfo?.tenantId;
      if (!tenantId) {
        setTagOptions([]);
        return;
      }

      listTags({
        tenantId,
        scope: 'logic_graph',
        pageSize: 100,
      })
        .then((response) => {
          if (response.code === 0 && response.data) {
            setTagOptions(
              response.data.list?.map((tag) => ({
                label: tag.name,
                value: tag.id,
              })) || [],
            );
          } else {
            console.error('加载标签列表失败:', response.msg || response.message);
            setTagOptions([]);
          }
        })
        .catch((error) => {
          console.error('加载标签列表失败:', error);
          setTagOptions([]);
        });
    } else {
      // 关闭时清空标签选项
      setTagOptions([]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editModalOpen]);

  if (!data) {
    return <Empty description="暂无数据" />;
  }

  // 获取图标组件
  const IconComponent = data.icon ? getIconComponent(data.icon) : null;
  const iconColor = data.iconColor || getIconColor(data.icon);

  return (
    <>
      <Card
        title="基本信息"
        loading={loading}
        extra={
          <Button type="primary" icon={<EditOutlined />} onClick={() => setEditModalOpen(true)}>
            编辑
          </Button>
        }
      >
        <Descriptions column={2} bordered>
          <Descriptions.Item label="逻辑图ID" span={2}>
            <Text copyable>{data.id}</Text>
          </Descriptions.Item>

          <Descriptions.Item label="图名称">{data.name}</Descriptions.Item>

          <Descriptions.Item label="版本号">
            <Tag color="blue">{data.version}</Tag>
          </Descriptions.Item>

          <Descriptions.Item label="启用状态" span={2}>
            <Badge
              status={data.isEnabled ? 'success' : 'default'}
              text={data.isEnabled ? '已启用' : '已禁用'}
            />
          </Descriptions.Item>

          <Descriptions.Item label="图标">
            {IconComponent ? (
              <Space>
                <IconComponent size={24} color={iconColor} />
                <Text type="secondary">{data.icon}</Text>
              </Space>
            ) : (
              <Text type="secondary">-</Text>
            )}
          </Descriptions.Item>

          <Descriptions.Item label="标签" span={2}>
            {data.tags && data.tags.length > 0 ? (
              <Space wrap>
                {/* Phase 2.5: 改为使用对象形式（tag.id, tag.name） */}
                {data.tags.map((tag) => (
                  <Tag key={tag.id} color="blue">
                    {tag.name}
                  </Tag>
                ))}
              </Space>
            ) : (
              <Text type="secondary">-</Text>
            )}
          </Descriptions.Item>

          <Descriptions.Item label="描述" span={2}>
            {data.description || <Text type="secondary">-</Text>}
          </Descriptions.Item>

          <Descriptions.Item label="创建人">
            {loadingUserNames ? (
              <Text type="secondary">加载中...</Text>
            ) : createdByName ? (
              createdByName
            ) : (
              <Text type="secondary">-</Text>
            )}
          </Descriptions.Item>

          <Descriptions.Item label="创建时间">
            {data.createdAt ? dayjs(data.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>

          <Descriptions.Item label="更新人">
            {loadingUserNames ? (
              <Text type="secondary">加载中...</Text>
            ) : updatedByName ? (
              updatedByName
            ) : (
              <Text type="secondary">-</Text>
            )}
          </Descriptions.Item>

          <Descriptions.Item label="更新时间">
            {data.updatedAt ? dayjs(data.updatedAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 编辑Modal */}
      <ModalForm
        title="编辑基本信息"
        open={editModalOpen}
        onOpenChange={(visible) => {
          if (!visible) {
            // 关闭时重置表单
            formRef.current?.resetFields();
            setTagOptions([]);
          }
          setEditModalOpen(visible);
        }}
        formRef={formRef}
        onFinish={async (values) => {
          const success = await onUpdate(values);
          if (success) {
            setEditModalOpen(false);
          }
          return success;
        }}
        modalProps={{
          destroyOnClose: true,
        }}
        initialValues={{
          name: data.name,
          version: data.version,
          description: data.description,
          tagIds: data.tags?.map(tag => tag.id) || [],
          isEnabled: data.isEnabled,
          icon: data.icon,
        }}
      >
        <ProFormText
          name="name"
          label="图名称"
          rules={[{ required: true, message: '请输入图名称' }]}
        />

        <ProFormText
          name="version"
          label="版本号"
          rules={[{ required: true, message: '请输入版本号' }]}
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
          fieldProps={{ rows: 4 }}
          placeholder="请输入逻辑图描述"
        />

        <ProFormSelect
          name="tagIds"
          label="标签"
          mode="multiple"
          placeholder="请选择标签"
          tooltip="标签用于分类和筛选，支持多选"
          options={tagOptions}
          fieldProps={{
            showSearch: true,
            optionFilterProp: 'label',
          }}
        />

        <ProFormSwitch name="isEnabled" label="启用状态" />
      </ModalForm>
    </>
  );
};

export default BasicInfoTab;
