/**
 * 动态积木配置表单组件
 * @description 根据 BlockSpec 的 configSchema 动态生成表单
 */

import {
  CloseCircleOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import type { FormInstance } from "antd";
import {
  Alert,
  Button,
  Empty,
  Form,
  Input,
  InputNumber,
  Select,
  Switch,
  TimePicker,
  Tooltip,
  Typography,
} from "antd";
import dayjs from "dayjs";
import { parseFirst } from "pgsql-ast-parser";
import React, { useEffect, useMemo } from "react";
import type { InfoAtomType } from "@/services/lynxmanager/types";
import styles from "./style.less";

interface ConfigField {
  type?: string;
  title?: string;
  description?: string;
  placeholder?: string;
  default?: any;
  enum?: any[];
  enumLabels?: string[]; // 与 enum 对应的中文标签
  check?: string; // 后端校验标记：must 表示必填
  properties?: Record<string, ConfigField>;
  items?: ConfigField;
  required?: string[];
  source?: string; // 数据源：infoAtomFields 表示从信息原子字段列表获取
  format?: string; // 格式校验：sql 表示 SQL 查询
  itemAddable?: boolean; // 数组项是否支持动态添加（渲染为列表而非逗号分隔）
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
  /** 信息原子类型列表（用于 source: "infoAtomFields" 的字段） */
  infoAtomTypes?: InfoAtomType[];
  /** 当前节点订阅的信息原子类型 ID 列表 */
  selectedInfoAtomTypeIds?: string[];
}

/**
 * 动态积木配置表单
 */
const DynamicBlockConfigForm: React.FC<Props> = ({
  schema,
  form,
  initialValues,
  onChange,
  infoAtomTypes = [],
  selectedInfoAtomTypeIds = [],
}) => {
  const [internalForm] = Form.useForm();
  const activeForm = form?.current || internalForm;

  // 从选中的信息原子类型中提取所有字段名（用于 source: "infoAtomFields"）
  const infoAtomFieldOptions = useMemo(() => {
    const fieldSet = new Set<string>();

    // 获取选中的信息原子类型
    const selectedTypes = infoAtomTypes.filter(
      (t) => selectedInfoAtomTypeIds.includes(t.id)
    );

    // 提取所有字段
    selectedTypes.forEach((type) => {
      type.dataFormat?.fields?.forEach((field) => {
        if (field.fieldKey) {
          fieldSet.add(field.fieldKey);
        }
      });
    });

    return Array.from(fieldSet).map((field) => ({
      label: field,
      value: field,
    }));
  }, [infoAtomTypes, selectedInfoAtomTypeIds]);

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
      // 支持两种必填标记：JSON Schema 的 required 数组，或后端的 check: "must"
      isRequired: schema.required?.includes(fieldName) || (fieldSchema as ConfigField).check === "must",
    }));
  }, [schema]);

  if (!schema || !formFields.length) {
    return <Empty description="暂无配置项" />;
  }

  // 渲染标签
  const renderLabel = (
    fieldSchema: ConfigField,
    fieldName: string,
    isRequired: boolean
  ) => (
    <span>
      {fieldSchema.title || fieldName}
      {isRequired && <span className={styles.requiredMark}>*</span>}
      {fieldSchema.description && (
        <Tooltip
          title={fieldSchema.description}
          placement="left"
          styles={{ root: { maxWidth: 300 } }}
          align={{ offset: [-10, 0] }}
        >
          <InfoCircleOutlined className={styles.descriptionIcon} />
        </Tooltip>
      )}
    </span>
  );

  // 判断字段是否为时间格式（HH:MM）
  const isTimeField = (propName: string, propSchema: ConfigField): boolean => {
    const desc = propSchema.description?.toLowerCase() || "";
    const title = propSchema.title?.toLowerCase() || "";
    const name = propName.toLowerCase();
    return (
      desc.includes("hh:mm") ||
      name.includes("time") ||
      title.includes("时间") ||
      name === "starttime" ||
      name === "endtime"
    );
  };

  // SQL 校验规则（使用 pgsql-ast-parser 进行完整语法校验）
  const sqlValidator = async (_: unknown, value: string) => {
    if (!value || value.trim() === "") {
      return Promise.resolve();
    }

    const trimmed = value.trim().toUpperCase();

    // 1. 基本检查：必须以 SELECT 开头
    if (!trimmed.startsWith("SELECT")) {
      return Promise.reject(new Error("仅支持 SELECT 查询语句"));
    }

    // 2. 检查危险关键字（防止 SQL 注入）
    const dangerousKeywords = ["INSERT", "UPDATE", "DELETE", "DROP", "TRUNCATE", "ALTER", "CREATE"];
    for (const keyword of dangerousKeywords) {
      const regex = new RegExp(`\\b${keyword}\\b`, "i");
      if (regex.test(value)) {
        return Promise.reject(new Error(`不允许使用 ${keyword} 语句`));
      }
    }

    // 3. PostgreSQL 语法校验
    try {
      parseFirst(value);
      return Promise.resolve();
    } catch (e) {
      // 提取错误信息，移除过于技术性的细节
      let msg = "SQL 语法错误";
      if (e instanceof Error) {
        // pgsql-ast-parser 的错误信息格式：'Syntax error at line X col Y: ...'
        const match = e.message.match(/at line (\d+) col (\d+)/);
        if (match) {
          msg = `语法错误（第 ${match[1]} 行，第 ${match[2]} 列）`;
        } else if (e.message.includes("Unexpected")) {
          msg = `语法错误：${e.message.replace(/^.*?Unexpected/, "意外的")}`;
        }
      }
      return Promise.reject(new Error(msg));
    }
  };

  // 渲染字符串数组的动态列表（itemAddable: true）
  const renderStringArrayList = (
    fieldName: string,
    fieldSchema: ConfigField,
    isRequired: boolean
  ) => {
    const label = renderLabel(fieldSchema, fieldName, isRequired);
    const itemPlaceholder = fieldSchema.items?.placeholder || `输入${fieldSchema.title || fieldName}`;

    // 数组必填校验规则
    const arrayRules = isRequired
      ? [
          {
            validator: async (_: unknown, value: unknown[]) => {
              if (!value || value.length === 0) {
                return Promise.reject(new Error(`请至少添加一个${fieldSchema.title || fieldName}`));
              }
              return Promise.resolve();
            },
          },
        ]
      : [];

    return (
      <Form.Item key={fieldName} label={label} rules={arrayRules}>
        <Form.List name={fieldName}>
          {(fields, { add, remove }) => (
            <div className={styles.stringArrayContainer}>
              {fields.length === 0 && (
                <div className={styles.arrayListEmpty}>
                  <Typography.Text type="secondary">
                    暂无参数，点击下方按钮添加
                  </Typography.Text>
                </div>
              )}
              {fields.map(({ key, name, ...restField }, index) => (
                <div key={key} className={styles.stringArrayItem}>
                  <span className={styles.stringArrayIndex}>${index + 1}</span>
                  <Form.Item
                    {...restField}
                    name={name}
                    style={{ marginBottom: 0, flex: 1 }}
                    rules={[{ required: true, message: "请输入参数值" }]}
                  >
                    <Input placeholder={itemPlaceholder} allowClear />
                  </Form.Item>
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => remove(name)}
                  />
                </div>
              ))}
              <Button
                type="dashed"
                onClick={() => add("")}
                block
                icon={<PlusOutlined />}
                className={styles.stringArrayAddButton}
              >
                添加参数
              </Button>
            </div>
          )}
        </Form.List>
      </Form.Item>
    );
  };

  // 渲染对象数组的动态列表
  const renderObjectArrayList = (
    fieldName: string,
    fieldSchema: ConfigField,
    isRequired: boolean
  ) => {
    const itemProperties = fieldSchema.items?.properties || {};
    const propEntries = Object.entries(itemProperties);

    if (propEntries.length === 0) {
      // 没有定义 properties，回退到 JSON 编辑器
      return renderJsonEditor(fieldName, fieldSchema, isRequired);
    }

    const label = renderLabel(fieldSchema, fieldName, isRequired);

    // 数组必填校验规则
    const arrayRules = isRequired
      ? [
          {
            validator: async (_: unknown, value: unknown[]) => {
              if (!value || value.length === 0) {
                return Promise.reject(new Error(`请至少添加一个${fieldSchema.title || fieldName}`));
              }
              return Promise.resolve();
            },
          },
        ]
      : [];

    return (
      <Form.Item key={fieldName} label={label} rules={arrayRules}>
        <Form.List name={fieldName}>
          {(fields, { add, remove }) => (
            <div className={styles.arrayListContainer}>
              {fields.length === 0 && (
                <div className={styles.arrayListEmpty}>
                  <Typography.Text type="secondary">
                    暂无配置项，点击下方按钮添加
                  </Typography.Text>
                </div>
              )}
              {fields.map(({ key, name, ...restField }) => (
                <div key={key} className={styles.arrayListItem}>
                  <div className={styles.arrayListItemContent}>
                    {propEntries.map(([propName, propSchema]) => {
                      const propField = propSchema as ConfigField;
                      const isTime = isTimeField(propName, propField);
                      // 支持两种必填标记：required 数组或 check: "must"
                      const isPropRequired =
                        fieldSchema.items?.required?.includes(propName) || propField.check === "must";

                      return (
                        <div key={propName} className={styles.arrayListItemField}>
                          <Form.Item
                            {...restField}
                            name={[name, propName]}
                            label={propField.title || propName}
                            rules={
                              isPropRequired
                                ? [{ required: true, message: `请输入${propField.title || propName}` }]
                                : []
                            }
                            style={{ marginBottom: 0 }}
                            // 时间字段需要特殊处理值
                            getValueProps={isTime ? (value) => ({
                              value: value ? dayjs(value, "HH:mm") : undefined,
                            }) : undefined}
                            getValueFromEvent={isTime ? (time) =>
                              time ? time.format("HH:mm") : undefined
                            : undefined}
                          >
                            {isTime ? (
                              <TimePicker
                                format="HH:mm"
                                placeholder={propField.placeholder || `选择${propField.title || propName}`}
                                style={{ width: "100%" }}
                              />
                            ) : (
                              <Input
                                placeholder={propField.placeholder || `输入${propField.title || propName}`}
                                allowClear
                              />
                            )}
                          </Form.Item>
                        </div>
                      );
                    })}
                  </div>
                  <div className={styles.arrayListItemActions}>
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => remove(name)}
                    />
                  </div>
                </div>
              ))}
              <div className={styles.arrayListAddButton}>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  添加{fieldSchema.items?.title || fieldSchema.title || "项目"}
                </Button>
              </div>
            </div>
          )}
        </Form.List>
      </Form.Item>
    );
  };

  // 渲染 JSON 编辑器
  const renderJsonEditor = (
    fieldName: string,
    fieldSchema: ConfigField,
    isRequired: boolean
  ) => {
    const label = renderLabel(fieldSchema, fieldName, isRequired);
    const rules = isRequired
      ? [{ required: true, message: `请输入${fieldSchema.title || fieldName}` }]
      : [];

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
  };

  // 根据字段类型渲染表单组件
  const renderField = (
    fieldName: string,
    fieldSchema: ConfigField,
    isRequired: boolean
  ) => {
    const fieldType = fieldSchema.type || "string";
    const label = renderLabel(fieldSchema, fieldName, isRequired);

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
                  const displayLabel =
                    v === "info" ? "信息" : v === "error" ? "错误" : v;

                  if (v === "info") {
                    return {
                      label: (
                        <span className={styles.levelOption}>
                          <InfoCircleOutlined className="icon-info" />
                          <span>{displayLabel}</span>
                        </span>
                      ),
                      value: v,
                    };
                  }
                  if (v === "error") {
                    return {
                      label: (
                        <span className={styles.levelOption}>
                          <CloseCircleOutlined className="icon-error" />
                          <span>{displayLabel}</span>
                        </span>
                      ),
                      value: v,
                    };
                  }
                  return { label: displayLabel, value: v };
                })
              : fieldSchema.enum.map((v, i) => ({
                  label: fieldSchema.enumLabels?.[i] ?? v,
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
        // 判断是否为时间字段（HH:MM 格式）
        if (isTimeField(fieldName, fieldSchema)) {
          return (
            <Form.Item
              key={fieldName}
              name={fieldName}
              label={label}
              rules={rules}
              getValueProps={(value) => ({
                value: value ? dayjs(value, "HH:mm") : undefined,
              })}
              getValueFromEvent={(time) =>
                time ? time.format("HH:mm") : undefined
              }
            >
              <TimePicker
                format="HH:mm"
                placeholder={fieldSchema.placeholder || `选择${fieldSchema.title || fieldName}`}
                style={{ width: "100%" }}
              />
            </Form.Item>
          );
        }
        // SQL 格式字段
        if (fieldSchema.format === "sql") {
          const sqlRules = [
            ...rules,
            { validator: sqlValidator, validateTrigger: ["onChange", "onBlur"] },
          ];
          return (
            <Form.Item
              key={fieldName}
              name={fieldName}
              label={label}
              rules={sqlRules}
              validateFirst
              validateTrigger={["onChange", "onBlur"]}
              className={styles.sqlEditorContainer}
            >
              <Input.TextArea
                placeholder={fieldSchema.placeholder || "SELECT ... FROM ..."}
                rows={4}
                className={styles.sqlEditor}
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
        // 如果 items 是对象类型，使用动态列表
        if (fieldSchema.items?.type === "object" && fieldSchema.items?.properties) {
          return renderObjectArrayList(fieldName, fieldSchema, isRequired);
        }

        // 如果 items 是字符串类型
        if (fieldSchema.items?.type === "string") {
          // 如果 itemAddable 为 true，使用动态添加列表
          if (fieldSchema.itemAddable) {
            return renderStringArrayList(fieldName, fieldSchema, isRequired);
          }
          // 如果 source 为 infoAtomFields，从信息原子字段列表获取选项
          if (fieldSchema.source === "infoAtomFields") {
            return (
              <Form.Item
                key={fieldName}
                name={fieldName}
                label={label}
                rules={rules}
                tooltip={infoAtomFieldOptions.length === 0 ? "请先在入口节点配置中选择订阅的信息原子类型" : undefined}
              >
                <Select
                  mode="multiple"
                  placeholder={infoAtomFieldOptions.length > 0
                    ? (fieldSchema.placeholder || "选择字段")
                    : "请先选择信息原子类型"}
                  options={infoAtomFieldOptions}
                  allowClear
                  disabled={infoAtomFieldOptions.length === 0}
                />
              </Form.Item>
            );
          }
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
                  options={fieldSchema.items.enum.map((v, i) => ({
                    label: fieldSchema.items?.enumLabels?.[i] ?? v,
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
        // 其他数组类型使用 JSON 编辑器
        return renderJsonEditor(fieldName, fieldSchema, isRequired);

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
        component={false}
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
