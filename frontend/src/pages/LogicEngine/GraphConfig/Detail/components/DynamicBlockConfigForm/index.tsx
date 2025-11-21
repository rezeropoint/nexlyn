/**
 * 动态积木配置表单组件
 * @description 根据 BlockSpec 的 configSchema 动态生成表单
 */

import { CloseCircleOutlined, InfoCircleOutlined } from "@ant-design/icons";
import type { FormInstance } from "antd";
import {
  Alert,
  Empty,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  theme,
  Tooltip,
} from "antd";
import React, { useEffect, useMemo } from "react";
import styles from "./style.less";

interface ConfigField {
  type?: string;
  title?: string;
  description?: string;
  placeholder?: string;
  default?: any;
  enum?: any[];
  properties?: Record<string, ConfigField>;
  items?: ConfigField;
  required?: string[];
}

interface ConfigSchema {
  type?: string;
  title?: string;
  description?: string;
  properties?: Record<string, ConfigField>;
  required?: string[];
}

interface Props {
  schema?: ConfigSchema;
  form: React.MutableRefObject<FormInstance | null>;
  initialValues?: Record<string, any>;
  onChange?: (values: Record<string, any>) => void;
}

/**
 * 动态积木配置表单
 */
const DynamicBlockConfigForm: React.FC<Props> = ({
  schema,
  form,
  initialValues,
  onChange,
}) => {
  const [internalForm] = Form.useForm();
  const activeForm = form?.current || internalForm;
  const { token } = theme.useToken();

  // 初始化表单值
  useEffect(() => {
    if (initialValues) {
      activeForm.setFieldsValue(initialValues);
    }
  }, [initialValues, activeForm]);

  // 解析 schema，生成表单字段
  // 支持两种格式：
  // 1. 标准格式：{ properties: {...}, required: [...] }
  // 2. 平铺格式：{ level: {...}, message: {...} }
  const formFields = useMemo(() => {
    if (!schema) {
      return [];
    }

    // 检查是否为标准 JSON Schema 格式（有 properties 字段）
    const properties = schema.properties
      ? schema.properties
      : // 否则将整个 schema 作为 properties（平铺格式）
        Object.fromEntries(
          Object.entries(schema).filter(
            ([key]) =>
              !["type", "required", "title", "description"].includes(key)
          )
        );

    if (!properties || Object.keys(properties).length === 0) {
      return [];
    }

    return Object.entries(properties).map(([fieldName, fieldSchema]) => ({
      name: fieldName,
      schema: fieldSchema as ConfigField,
      isRequired: schema.required?.includes(fieldName) || false,
    }));
  }, [schema]);

  if (!schema || !formFields.length) {
    return <Empty description="暂无配置项" />;
  }

  // 根据字段类型渲染表单组件
  const renderField = (
    fieldName: string,
    fieldSchema: ConfigField,
    isRequired: boolean
  ) => {
    const fieldType = fieldSchema.type || "string";
    const label = (
      <span>
        {fieldSchema.title || fieldName}
        {isRequired && (
          <span style={{ color: token.colorError, marginLeft: 4 }}>*</span>
        )}
        {fieldSchema.description && (
          <Tooltip
            title={fieldSchema.description}
            placement="left"
            overlayStyle={{ maxWidth: 300 }}
            align={{ offset: [-10, 0] }}
          >
            <InfoCircleOutlined
              style={{
                marginLeft: 8,
                cursor: "help",
                color: token.colorTextTertiary,
                flexShrink: 0,
              }}
            />
          </Tooltip>
        )}
      </span>
    );

    const rules = isRequired
      ? [
          {
            required: true,
            message: `请输入${fieldSchema.title || fieldName}`,
          },
        ]
      : [];

    switch (fieldType) {
      case "string":
        // 如果有 enum，使用 Select
        if (fieldSchema.enum && fieldSchema.enum.length > 0) {
          // 为特定字段名提供自定义渲染（如 level 字段）
          const customOptions =
            fieldName === "level"
              ? fieldSchema.enum.map((v) => {
                  // 为日志级别添加图标和颜色
                  let icon = null;
                  let color = token.colorTextSecondary;
                  const displayLabel =
                    v === "info" ? "信息" : v === "error" ? "错误" : v;

                  if (v === "info") {
                    icon = <InfoCircleOutlined />;
                    color = token.colorInfo;
                  } else if (v === "error") {
                    icon = <CloseCircleOutlined />;
                    color = token.colorError;
                  }

                  return {
                    label: (
                      <span
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 6,
                        }}
                      >
                        {icon && <span style={{ color }}>{icon}</span>}
                        <span>{displayLabel}</span>
                      </span>
                    ),
                    value: v,
                  };
                })
              : fieldSchema.enum.map((v) => ({
                  label: v,
                  value: v,
                }));

          return (
            <Form.Item
              key={fieldName}
              name={fieldName}
              label={label}
              rules={rules}
            >
              <Select
                placeholder={fieldSchema.placeholder || "请选择"}
                options={customOptions}
                allowClear
              />
            </Form.Item>
          );
        }
        // 否则使用 Input
        return (
          <Form.Item
            key={fieldName}
            name={fieldName}
            label={label}
            rules={rules}
          >
            <Input
              placeholder={
                fieldSchema.placeholder ||
                `输入${fieldSchema.title || fieldName}`
              }
              allowClear
            />
          </Form.Item>
        );

      case "number":
      case "integer":
        return (
          <Form.Item
            key={fieldName}
            name={fieldName}
            label={label}
            rules={rules}
            className={styles.fullWidthInput}
          >
            <InputNumber
              placeholder={
                fieldSchema.placeholder ||
                `输入${fieldSchema.title || fieldName}`
              }
            />
          </Form.Item>
        );

      case "boolean":
        return (
          <Form.Item
            key={fieldName}
            name={fieldName}
            label={label}
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
        );

      case "array":
        // 如果 items 是字符串类型，使用 Select 多选
        if (fieldSchema.items?.type === "string") {
          // 如果有 enum 值，使用下拉选择
          if (fieldSchema.items.enum && fieldSchema.items.enum.length > 0) {
            return (
              <Form.Item
                key={fieldName}
                name={fieldName}
                label={label}
                rules={rules}
              >
                <Select
                  mode="multiple"
                  placeholder={fieldSchema.placeholder || "选择项目"}
                  options={fieldSchema.items.enum.map((v) => ({
                    label: v,
                    value: v,
                  }))}
                  allowClear
                />
              </Form.Item>
            );
          }
          // 否则使用文本输入（多个用逗号分隔）
          return (
            <Form.Item
              key={fieldName}
              name={fieldName}
              label={label}
              rules={rules}
              tooltip="多个值用逗号分隔"
            >
              <Input.TextArea
                placeholder={fieldSchema.placeholder || "多个值用逗号分隔"}
                rows={2}
              />
            </Form.Item>
          );
        }
        // 否则使用 JSON 编辑器
        return (
          <Form.Item
            key={fieldName}
            name={fieldName}
            label={label}
            rules={rules}
            className={styles.jsonEditorContainer}
          >
            <Input.TextArea
              placeholder={fieldSchema.placeholder || "JSON 格式"}
              rows={4}
              className={styles.jsonEditor}
            />
          </Form.Item>
        );

      default:
        // 其他类型默认使用文本输入
        return (
          <Form.Item
            key={fieldName}
            name={fieldName}
            label={label}
            rules={rules}
          >
            <Input
              placeholder={
                fieldSchema.placeholder ||
                `输入${fieldSchema.title || fieldName}`
              }
            />
          </Form.Item>
        );
    }
  };

  // 检查是否存在需要至少填一项的字段组合（message、contextKey、printInfoAtom）
  const hasMultiSourceFields = formFields.some((f) =>
    ["message", "contextKey", "printInfoAtom"].includes(f.name)
  );

  return (
    <div className="dynamic-block-config-form">
      <Form
        form={activeForm}
        layout="vertical"
        onValuesChange={(_, values) => {
          onChange?.(values);
        }}
      >
        {formFields.map(({ name, schema: fieldSchema, isRequired }) =>
          renderField(name, fieldSchema, isRequired)
        )}

        {/* 配置验证提示 */}
        {hasMultiSourceFields && (
          <div className={styles.validationTip}>
            <Alert
              type="warning"
              showIcon
              message="配置提示"
              description="如果 message 为空，需至少配置 contextKey 或 printInfoAtom 中的一项以获取日志内容"
              className={styles.compactAlert}
            />
          </div>
        )}
      </Form>
    </div>
  );
};

export default DynamicBlockConfigForm;
