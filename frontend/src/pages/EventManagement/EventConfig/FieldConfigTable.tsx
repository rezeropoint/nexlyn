import {
  ArrowDownOutlined,
  ArrowUpOutlined,
  DeleteOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import { Button, Input, InputNumber, Select, Switch, Table } from "antd";
import React from "react";
import { FIELD_TYPE_OPTIONS } from "../constants";
import type { FieldConfigInput, FieldMetadata } from "../types";
import styles from "./FieldConfigTable.less";

interface FieldConfigTableProps {
  fields: FieldConfigInput[];
  onChange: (fields: FieldConfigInput[]) => void;
  availableFields: FieldMetadata[];
}

/**
 * 字段配置表格
 */
const FieldConfigTable: React.FC<FieldConfigTableProps> = ({
  fields,
  onChange,
  availableFields,
}) => {
  // 添加字段
  const handleAdd = () => {
    const newField: FieldConfigInput = {
      fieldName: "",
      displayName: "",
      fieldType: "string",
      isVisible: true,
      displayOrder: fields.length,
      isSearchable: false,
    };
    onChange([...fields, newField]);
  };

  // 删除字段
  const handleDelete = (index: number) => {
    const newFields = fields.filter((_, i) => i !== index);
    // 重新计算显示顺序
    newFields.forEach((field, i) => {
      field.displayOrder = i;
    });
    onChange(newFields);
  };

  // 更新字段
  const handleUpdate = (
    index: number,
    key: keyof FieldConfigInput,
    value: any
  ) => {
    const newFields = [...fields];
    newFields[index] = {
      ...newFields[index],
      [key]: value,
    };
    onChange(newFields);
  };

  // 移动字段位置
  const moveField = (index: number, direction: "up" | "down") => {
    const newIndex = direction === "up" ? index - 1 : index + 1;
    if (newIndex < 0 || newIndex >= fields.length) return;

    const newFields = [...fields];
    const temp = newFields[index];
    newFields[index] = newFields[newIndex];
    newFields[newIndex] = temp;

    // 重新计算显示顺序
    newFields.forEach((field, i) => {
      field.displayOrder = i;
    });
    onChange(newFields);
  };

  // 表格列定义
  const columns = [
    {
      title: "排序",
      dataIndex: "sort",
      width: 100,
      render: (_: any, __: any, index: number) => (
        <div>
          <Button
            type="text"
            size="small"
            icon={<ArrowUpOutlined />}
            disabled={index === 0}
            onClick={() => moveField(index, "up")}
          />
          <Button
            type="text"
            size="small"
            icon={<ArrowDownOutlined />}
            disabled={index === fields.length - 1}
            onClick={() => moveField(index, "down")}
          />
        </div>
      ),
    },
    {
      title: "字段名",
      dataIndex: "fieldName",
      width: 200,
      render: (_: any, record: FieldConfigInput, index: number) => (
        <Select
          value={record.fieldName}
          onChange={(value) => {
            handleUpdate(index, "fieldName", value);
            // 如果显示名称为空，自动填充字段名
            if (!record.displayName) {
              handleUpdate(index, "displayName", value);
            }
          }}
          placeholder="选择字段"
          className={styles.fullWidthSelect}
        >
          {availableFields.map((field) => (
            <Select.Option key={field.fieldName} value={field.fieldName}>
              {field.fieldName}
            </Select.Option>
          ))}
        </Select>
      ),
    },
    {
      title: "显示名称",
      dataIndex: "displayName",
      width: 200,
      render: (_: any, record: FieldConfigInput, index: number) => (
        <Input
          value={record.displayName}
          onChange={(e) => handleUpdate(index, "displayName", e.target.value)}
          placeholder="输入显示名称"
        />
      ),
    },
    {
      title: "字段类型",
      dataIndex: "fieldType",
      width: 150,
      render: (_: any, record: FieldConfigInput, index: number) => (
        <Select
          value={record.fieldType}
          onChange={(value) => handleUpdate(index, "fieldType", value)}
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
      title: "显示顺序",
      dataIndex: "displayOrder",
      width: 100,
      render: (_: any, record: FieldConfigInput, index: number) => (
        <InputNumber
          value={record.displayOrder}
          onChange={(value) => handleUpdate(index, "displayOrder", value || 0)}
          min={0}
          className={styles.fullWidthInput}
        />
      ),
    },
    {
      title: "列表显示",
      dataIndex: "isVisible",
      width: 100,
      align: "center" as const,
      render: (_: any, record: FieldConfigInput, index: number) => (
        <Switch
          checked={record.isVisible}
          onChange={(checked) => handleUpdate(index, "isVisible", checked)}
        />
      ),
    },
    {
      title: "可搜索",
      dataIndex: "isSearchable",
      width: 100,
      align: "center" as const,
      render: (_: any, record: FieldConfigInput, index: number) => (
        <Switch
          checked={record.isSearchable}
          onChange={(checked) => handleUpdate(index, "isSearchable", checked)}
        />
      ),
    },
    {
      title: "操作",
      width: 80,
      align: "center" as const,
      render: (_: any, __: any, index: number) => (
        <Button
          type="text"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => handleDelete(index)}
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
        rowKey={(_, index) => index?.toString() || "0"}
        pagination={false}
        size="small"
      />
    </div>
  );
};

export default FieldConfigTable;
