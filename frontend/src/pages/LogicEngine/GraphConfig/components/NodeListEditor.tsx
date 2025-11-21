/**
 * 节点列表编辑器组件
 */

import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Select, Space, Spin, Switch, Table, Tooltip } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useCallback, useEffect, useState } from 'react';
import { getBlockSpecs, type BlockSpec } from '@/services/lynxmanager';
import { DEFAULT_NODE, NODE_TYPE_OPTIONS } from '../constants';
import type { NodeItem } from '../types';
import styles from './GraphConfigForm.less';

interface NodeListEditorProps {
  nodes: NodeItem[];
  onChange: (nodes: NodeItem[]) => void;
}

/**
 * 节点列表编辑器
 */
const NodeListEditor: React.FC<NodeListEditorProps> = ({ nodes, onChange }) => {
  // 逻辑块规格列表
  const [blockSpecs, setBlockSpecs] = useState<BlockSpec[]>([]);
  const [loading, setLoading] = useState(false);

  // 获取逻辑块规格列表
  useEffect(() => {
    setLoading(true);
    getBlockSpecs()
      .then((response) => {
        if (response.code === 0 && response.data?.list) {
          setBlockSpecs(response.data.list);
        }
      })
      .catch((error) => {
        console.error('获取逻辑块规格列表失败:', error);
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  // 添加节点
  const handleAdd = useCallback(() => {
    // 生成临时ID（格式: client:node-{timestamp}-{random}）
    const tempId = `client:node-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
    const newNode: NodeItem = {
      ...DEFAULT_NODE,
      id: tempId,
      key: tempId,
    };
    onChange([...nodes, newNode]);
  }, [nodes, onChange]);

  // 删除节点
  const handleDelete = useCallback(
    (index: number) => {
      const newNodes = [...nodes];
      newNodes.splice(index, 1);
      onChange(newNodes);
    },
    [nodes, onChange],
  );

  // 更新节点字段
  const handleFieldChange = useCallback(
    (index: number, field: keyof NodeItem, value: any) => {
      const newNodes = [...nodes];
      newNodes[index] = { ...newNodes[index], [field]: value };
      onChange(newNodes);
    },
    [nodes, onChange],
  );

  // 表格列定义
  const columns: ColumnsType<NodeItem> = [
    {
      title: '节点类型',
      dataIndex: 'type',
      width: 150,
      render: (text, _record, index) => (
        <Select
          style={{ width: '100%' }}
          placeholder="请选择节点类型"
          value={text}
          onChange={(value) => handleFieldChange(index, 'type', value)}
          options={NODE_TYPE_OPTIONS}
        />
      ),
    },
    {
      title: (
        <Tooltip title="逻辑块的类型标识">
          <span>逻辑块</span>
        </Tooltip>
      ),
      dataIndex: 'blockType',
      width: 250,
      render: (text, record, index) => {
        // 根据当前节点的blockType和blockVersion，构建唯一选项值
        const currentValue = text && record.blockVersion ? `${text}::${record.blockVersion}` : '';

        // 构建逻辑块选项（blockType::version作为值）
        const blockOptions = blockSpecs.map((spec) => ({
          label: `${spec.name || spec.blockType} (${spec.version})`,
          value: `${spec.blockType}::${spec.version}`,
          spec,
        }));

        return (
          <Select
            style={{ width: '100%' }}
            placeholder="请选择逻辑块"
            value={currentValue || undefined}
            onChange={(value) => {
              // 解析选中的值（格式: blockType::version）
              const [blockType, version] = value.split('::');
              // 一次性更新两个字段，避免状态更新冲突
              const newNodes = [...nodes];
              newNodes[index] = {
                ...newNodes[index],
                blockType,
                blockVersion: version,
              };
              onChange(newNodes);
            }}
            options={blockOptions}
            showSearch
            optionFilterProp="label"
            loading={loading}
          />
        );
      },
    },
    {
      title: '入口节点',
      dataIndex: 'isEntryPoint',
      width: 100,
      render: (checked, _record, index) => (
        <Switch
          checked={checked}
          onChange={(checked) => handleFieldChange(index, 'isEntryPoint', checked)}
        />
      ),
    },
    {
      title: '操作',
      width: 80,
      fixed: 'right',
      render: (_, _record, index) => (
        <Button
          type="link"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => handleDelete(index)}
        >
          删除
        </Button>
      ),
    },
  ];

  return (
    <Spin spinning={loading} tip="加载逻辑块列表...">
      <Space direction="vertical" size="middle" className={styles.editorContainer}>
        <Table
          columns={columns}
          dataSource={nodes.map((node, index) => ({ ...node, key: node.key || `node-${index}` }))}
          pagination={false}
          scroll={{ x: 800 }}
          size="small"
          locale={{ emptyText: '暂无节点，请点击下方按钮添加' }}
        />
        <Button type="dashed" icon={<PlusOutlined />} onClick={handleAdd} block>
          添加节点
        </Button>
      </Space>
    </Spin>
  );
};

export default NodeListEditor;
