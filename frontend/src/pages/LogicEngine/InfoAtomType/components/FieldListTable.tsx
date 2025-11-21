/**
 * 字段列表表格组件
 */

import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  DeleteOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { Button, Input, Select, Table } from 'antd';
import React from 'react';
import { FIELD_TYPE_OPTIONS } from '../constants';
import styles from './FieldListTable.less';

export interface FieldItem {
  fieldKey: string;
  fieldPath: string;
  fieldType: string;
}

interface FieldListTableProps {
  fields: FieldItem[];
  onChange: (fields: FieldItem[]) => void;
}

/**
 * 字段列表表格
 */
const FieldListTable: React.FC<FieldListTableProps> = ({ fields, onChange }) => {
  // 添加字段
  const handleAdd = () => {
    const newField: FieldItem = {
      fieldKey: '',
      fieldPath: '',
      fieldType: 'string',
    };
    onChange([...fields, newField]);
  };

  // 删除字段
  const handleDelete = (index: number) => {
    const newFields = fields.filter((_, i) => i !== index);
    onChange(newFields);
  };

  // 更新字段
  const handleUpdate = (index: number, key: keyof FieldItem, value: any) => {
    const newFields = [...fields];
    newFields[index] = {
      ...newFields[index],
      [key]: value,
    };
    onChange(newFields);
  };

  // 移动字段位置
  const moveField = (index: number, direction: 'up' | 'down') => {
    const newIndex = direction === 'up' ? index - 1 : index + 1;
    if (newIndex < 0 || newIndex >= fields.length) return;

    const newFields = [...fields];
    const temp = newFields[index];
    newFields[index] = newFields[newIndex];
    newFields[newIndex] = temp;
    onChange(newFields);
  };

  // 表格列定义
  const columns = [
    {
      title: '排序',
      dataIndex: 'sort',
      width: 80,
      render: (_: any, __: any, index: number) => (
        <div>
          <Button
            type="text"
            size="small"
            icon={<ArrowUpOutlined />}
            disabled={index === 0}
            onClick={() => moveField(index, 'up')}
          />
          <Button
            type="text"
            size="small"
            icon={<ArrowDownOutlined />}
            disabled={index === fields.length - 1}
            onClick={() => moveField(index, 'down')}
          />
        </div>
      ),
    },
    {
      title: '字段键',
      dataIndex: 'fieldKey',
      width: 200,
      render: (_: any, record: FieldItem, index: number) => (
        <Input
          value={record.fieldKey}
          onChange={(e) => handleUpdate(index, 'fieldKey', e.target.value)}
          placeholder="请输入字段键"
        />
      ),
    },
    {
      title: '字段路径',
      dataIndex: 'fieldPath',
      width: 280,
      render: (_: any, record: FieldItem, index: number) => (
        <Input
          value={record.fieldPath}
          onChange={(e) => handleUpdate(index, 'fieldPath', e.target.value)}
          placeholder="请输入字段路径（支持JSONPath）"
        />
      ),
    },
    {
      title: '字段类型',
      dataIndex: 'fieldType',
      width: 150,
      render: (_: any, record: FieldItem, index: number) => (
        <Select
          value={record.fieldType}
          onChange={(value) => handleUpdate(index, 'fieldType', value)}
          className={styles.fullWidthSelect}
        >
          {FIELD_TYPE_OPTIONS.map((option) => (
            <Select.Option key={option.value} value={option.value}>
              {option.label}
            </Select.Option>
          ))}
        </Select>
      ),
    },
    {
      title: '操作',
      width: 60,
      align: 'center' as const,
      render: (_: any, __: any, index: number) => (
        <Button
          type="text"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => handleDelete(index)}
          disabled={fields.length === 1}
        />
      ),
    },
  ];

  return (
    <div className={styles.tableContainer}>
      <div className={styles.addButtonWrapper}>
        <Button type="dashed" onClick={handleAdd} icon={<PlusOutlined />} block>
          添加字段
        </Button>
      </div>
      <Table
        dataSource={fields}
        columns={columns}
        rowKey={(_, index) => index?.toString() || '0'}
        pagination={false}
        size="small"
      />
    </div>
  );
};

export default FieldListTable;
