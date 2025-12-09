/**
 * 边列表编辑器组件
 */

import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Input, Popconfirm, Select, Space, Table, Tooltip } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import React, { useCallback } from 'react';
import { DEFAULT_EDGE } from '../constants';
import type { EdgeItem, NodeItem } from '../types';
import styles from './EdgeListEditor.module.less';

interface EdgeListEditorProps {
  edges: EdgeItem[];
  nodes: NodeItem[]; // 用于节点选择
  onChange: (edges: EdgeItem[]) => void;
}

/**
 * 边列表编辑器
 */
const EdgeListEditor: React.FC<EdgeListEditorProps> = ({ edges, nodes, onChange }) => {
  // 节点选项
  const nodeOptions = nodes.map((node) => ({
    label: `${node.id || '未命名'} (${node.type || '未知类型'})`,
    value: node.id,
  }));

  // 添加边
  const handleAdd = useCallback(() => {
    // 生成临时ID（格式: client:edge-{timestamp}-{random}）
    const tempId = `client:edge-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
    const newEdge: EdgeItem = {
      ...DEFAULT_EDGE,
      id: tempId,
      key: tempId,
    };
    onChange([...edges, newEdge]);
  }, [edges, onChange]);

  // 删除边
  const handleDelete = useCallback(
    (index: number) => {
      const newEdges = [...edges];
      newEdges.splice(index, 1);
      onChange(newEdges);
    },
    [edges, onChange],
  );

  // 更新边字段
  const handleFieldChange = useCallback(
    (index: number, field: keyof EdgeItem, value: any) => {
      const newEdges = [...edges];
      newEdges[index] = { ...newEdges[index], [field]: value };
      onChange(newEdges);
    },
    [edges, onChange],
  );

  // 表格列定义
  const columns: ColumnsType<EdgeItem> = [
    {
      title: '源节点',
      dataIndex: 'sourceID',
      width: 200,
      render: (text, _record, index) => (
        <Select
          className={styles.fullWidth}
          placeholder="请选择源节点"
          value={text}
          onChange={(value) => handleFieldChange(index, 'sourceID', value)}
          options={nodeOptions}
          showSearch
          filterOption={(input, option) =>
            (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
          }
        />
      ),
    },
    {
      title: '目标节点',
      dataIndex: 'targetID',
      width: 200,
      render: (text, _record, index) => (
        <Select
          className={styles.fullWidth}
          placeholder="请选择目标节点"
          value={text}
          onChange={(value) => handleFieldChange(index, 'targetID', value)}
          options={nodeOptions}
          showSearch
          filterOption={(input, option) =>
            (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
          }
        />
      ),
    },
    {
      title: (
        <Tooltip title="支持变量：context.xxx.yyy（图上下文）、atom.xxx（信息原子字段）">
          条件表达式
        </Tooltip>
      ),
      dataIndex: 'condition',
      width: 240,
      render: (text, _record, index) => (
        <Input.TextArea
          placeholder='示例: context.time_window.window == "before_work"'
          value={text}
          onChange={(e) => handleFieldChange(index, 'condition', e.target.value)}
          rows={1}
          autoSize={{ minRows: 1, maxRows: 3 }}
          className={styles.conditionInput}
        />
      ),
    },
    {
      title: '操作',
      width: 80,
      fixed: 'right',
      render: (_, _record, index) => (
        <Popconfirm
          title="确认删除此边？"
          onConfirm={() => handleDelete(index)}
          okText="删除"
          cancelText="取消"
        >
          <Button type="link" danger size="small" icon={<DeleteOutlined />}>
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div>
      <Space direction="vertical" size="middle" className={styles.editorContainer}>
        <Table
          columns={columns}
          dataSource={edges.map((edge, index) => ({ ...edge, key: edge.key || `edge-${index}` }))}
          pagination={false}
          scroll={{ x: 800 }}
          size="small"
          locale={{ emptyText: '暂无边，请点击下方按钮添加' }}
        />
        <Button type="dashed" icon={<PlusOutlined />} onClick={handleAdd} block>
          添加边
        </Button>
      </Space>
    </div>
  );
};

export default EdgeListEditor;
