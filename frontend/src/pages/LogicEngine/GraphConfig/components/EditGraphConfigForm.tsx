/**
 * 编辑逻辑图配置表单组件
 */

import { useModel } from '@@/exports';
import { getGraphConfig, type GraphConfigMetadata } from '@/services/lynxmanager';
import { listTags } from '@/services/lynxmanager/api';
import {
  DrawerForm,
  ProForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { Alert, App, Spin } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import type { EdgeItem, NodeItem } from '../types';
import EdgeListEditor from './EdgeListEditor';
import styles from './GraphConfigForm.less';
import IconSelector from './IconSelector';
import NodeListEditor from './NodeListEditor';

interface EditGraphConfigFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: any) => Promise<boolean>;
  currentRow?: GraphConfigMetadata;
}

/**
 * 编辑逻辑图配置表单
 */
const EditGraphConfigForm: React.FC<EditGraphConfigFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const { message } = App.useApp();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const formRef = useRef<ProFormInstance>();

  // 加载状态
  const [loading, setLoading] = useState(false);
  // 节点和边列表
  const [nodes, setNodes] = useState<NodeItem[]>([]);
  const [edges, setEdges] = useState<EdgeItem[]>([]);
  // 标签选项列表
  const [tagOptions, setTagOptions] = useState<Array<{ label: string; value: string }>>([]);

  // 加载标签列表
  useEffect(() => {
    if (open) {
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
  }, [open]);

  // 加载逻辑图详情
  useEffect(() => {
    if (open && currentRow?.id) {
      setLoading(true);
      getGraphConfig(currentRow.id)
        .then((response) => {
          if (response.code === 0 && response.data) {
            const { nodes: nodeList, edges: edgeList, tags, ...rest } = response.data;

            // 设置节点和边
            setNodes(nodeList || []);
            setEdges(edgeList || []);

            // 设置表单值
            formRef.current?.setFieldsValue({
              ...rest,
              tagIds: tags?.map((tag) => tag.id) || [],
            });
          } else {
            message.error(response.msg || response.message || '获取逻辑图详情失败');
            onOpenChange(false);
          }
        })
        .catch((error) => {
          message.error(`获取逻辑图详情失败: ${error.message || '未知错误'}`);
          onOpenChange(false);
        })
        .finally(() => {
          setLoading(false);
        });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, currentRow?.id]);

  // 表单提交处理
  const handleSubmit = async (values: any) => {
    if (!currentRow) {
      message.error('未选择逻辑图配置');
      return false;
    }

    // 验证节点列表
    if (nodes.length === 0) {
      message.error('请至少添加一个节点');
      return false;
    }

    // 验证入口节点
    const hasEntryPoint = nodes.some((node) => node.isEntryPoint);
    if (!hasEntryPoint) {
      message.error('请至少设置一个入口节点');
      return false;
    }

    // 收集节点ID列表（用于验证边的节点引用）
    const nodeIds = nodes.map((n) => n.id);

    // 验证节点字段
    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i];
      if (!node.blockType) {
        message.error(`第 ${i + 1} 个节点的逻辑块类型不能为空`);
        return false;
      }
      if (!node.blockVersion) {
        message.error(`第 ${i + 1} 个节点的逻辑块版本不能为空`);
        return false;
      }
    }

    // 验证边字段
    for (let i = 0; i < edges.length; i++) {
      const edge = edges[i];
      if (!edge.sourceID || !edge.targetID) {
        message.error(`第 ${i + 1} 条边的源节点和目标节点不能为空`);
        return false;
      }
      // 防止节点自己连接自己
      if (edge.sourceID === edge.targetID) {
        message.error(`第 ${i + 1} 条边不能连接自己（源节点和目标节点相同）`);
        return false;
      }
      // 验证节点存在性
      if (!nodeIds.includes(edge.sourceID)) {
        message.error(`第 ${i + 1} 条边的源节点 "${edge.sourceID}" 不存在`);
        return false;
      }
      if (!nodeIds.includes(edge.targetID)) {
        message.error(`第 ${i + 1} 条边的目标节点 "${edge.targetID}" 不存在`);
        return false;
      }
    }

    // 构建请求数据
    const requestData = {
      tenantId: currentRow.tenantId,
      name: values.name,
      version: values.version,
      description: values.description || '',
      tagIds: values.tagIds || [],
      isEnabled: values.isEnabled !== undefined ? values.isEnabled : true,
      icon: values.icon, // 图标名称
      nodes: nodes.map(({ key, ...rest }) => rest), // 移除key字段
      edges: edges.map(({ key, ...rest }) => rest), // 移除key字段
    };

    return await onFinish(requestData);
  };

  return (
    <DrawerForm
      title="编辑逻辑图配置"
      width={1200}
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置表单和列表
          formRef.current?.resetFields();
          setNodes([]);
          setEdges([]);
          setTagOptions([]);
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
    >
      <Spin spinning={loading}>
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
          options={tagOptions}
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

        <Alert
          message="节点配置"
          description="请配置逻辑图的节点列表。至少需要一个节点，且必须有一个入口节点。"
          type="info"
          className={styles.configAlert}
        />

        <ProForm.Item label="节点列表" required>
          <NodeListEditor nodes={nodes} onChange={setNodes} />
        </ProForm.Item>

        <Alert
          message="边配置"
          description="请配置逻辑图的边列表。边用于连接节点，定义执行流程。"
          type="info"
          className={styles.configAlert}
        />

        <ProForm.Item label="边列表">
          <EdgeListEditor edges={edges} nodes={nodes} onChange={setEdges} />
        </ProForm.Item>
      </Spin>
    </DrawerForm>
  );
};

export default EditGraphConfigForm;
