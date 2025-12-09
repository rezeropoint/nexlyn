/**
 * 边条件配置表单组件
 * @description 提供可视化的条件配置，替代手写表达式
 */

import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { ProFormSelect, ProFormText } from '@ant-design/pro-components';
import { Button, Card, Flex, Space, Typography } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import type { InfoAtomType } from '@/services/lynxmanager/types';
import styles from './index.module.less';

const { Text } = Typography;

/** OutputSchema 中的属性定义 */
interface OutputSchemaProperty {
  type: string;
  description?: string;
}

/** 上下文输出结构定义 */
export interface ContextOutputInfo {
  contextKey: string; // 上下文键名（如 time_window）
  blockType: string; // 积木类型
  outputSchema: {
    contextKey?: string; // 键名来源（如 config.resultKey）
    type?: string; // 类型：object / dynamic-keys
    keySource?: string; // 动态键来源（如 config.windows[].name）
    valueType?: string; // 动态键的值类型（如 boolean）
    properties?: Record<string, OutputSchemaProperty>; // 固定属性
    description?: string;
  } | null;
  // 动态键的实际值（从节点配置解析）
  dynamicKeys?: string[];
}

/** 单个条件配置 */
export interface ConditionItem {
  /** 条件ID（前端用于列表渲染） */
  id?: string;
  /** 变量来源: context(图上下文) / atom(信息原子) */
  source: 'context' | 'atom';
  /** 变量路径，如 time_window.window 或 status */
  path: string;
  /** 比较操作符 */
  operator: '==' | '!=' | '>' | '<' | '>=' | '<=' | 'contains';
  /** 比较值 */
  value: string;
}

/** 条件组配置（支持 AND/OR 组合） */
export interface ConditionConfig {
  /** 条件逻辑: and(所有条件都满足) / or(任一条件满足) */
  logic: 'and' | 'or';
  /** 条件列表 */
  conditions: ConditionItem[];
}

interface EdgeConditionFormProps {
  /** 当前条件配置（JSON字符串或结构化对象） */
  value?: string | ConditionConfig;
  /** 条件变更回调 */
  onChange?: (value: string) => void;
  /** 信息原子类型列表（用于 atom 字段选择） */
  infoAtomTypes?: InfoAtomType[];
  /** 上下文键列表（用于 context 路径选择） */
  contextKeys?: string[];
  /** 上下文输出结构信息（用于智能提示子字段和值类型） */
  contextOutputInfos?: ContextOutputInfo[];
}

/** 操作符选项 */
const OPERATOR_OPTIONS = [
  { label: '等于 (==)', value: '==' },
  { label: '不等于 (!=)', value: '!=' },
  { label: '大于 (>)', value: '>' },
  { label: '小于 (<)', value: '<' },
  { label: '大于等于 (>=)', value: '>=' },
  { label: '小于等于 (<=)', value: '<=' },
  { label: '包含', value: 'contains' },
];

/** 变量来源选项 */
const SOURCE_OPTIONS = [
  { label: '图上下文 (context)', value: 'context' },
  { label: '信息原子 (atom)', value: 'atom' },
];

/** 逻辑选项 */
const LOGIC_OPTIONS = [
  { label: '全部满足 (AND)', value: 'and' },
  { label: '任一满足 (OR)', value: 'or' },
];

/** 生成唯一ID */
const generateId = () => `cond-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`;

/** 创建默认条件项 */
const createDefaultCondition = (): ConditionItem => ({
  id: generateId(),
  source: 'context',
  path: '',
  operator: '==',
  value: '',
});

/** 默认条件配置 */
const DEFAULT_CONFIG: ConditionConfig = {
  logic: 'and',
  conditions: [],
};

/**
 * 解析条件配置
 * 支持：空值、JSON字符串、结构化对象
 */
const parseConfig = (value?: string | ConditionConfig): ConditionConfig => {
  if (!value) return { ...DEFAULT_CONFIG };
  if (typeof value === 'object') {
    // 确保每个条件都有 ID
    return {
      ...value,
      conditions: value.conditions.map((c) => ({ ...c, id: c.id || generateId() })),
    };
  }

  try {
    const parsed = JSON.parse(value);
    // 检查是否是结构化配置
    if (parsed.logic && Array.isArray(parsed.conditions)) {
      // 为每个条件生成 ID
      return {
        ...parsed,
        conditions: parsed.conditions.map((c: ConditionItem) => ({
          ...c,
          id: c.id || generateId(),
        })),
      };
    }
  } catch {
    // 不是JSON，可能是旧的表达式格式，返回空配置
  }

  return { ...DEFAULT_CONFIG };
};

/**
 * 将条件配置序列化为JSON字符串
 * 注意：序列化时排除 id 字段，后端不需要
 */
const serializeConfig = (config: ConditionConfig): string => {
  // 过滤掉不完整的条件，并排除 id 字段
  const validConditions = config.conditions
    .filter((c) => c.path && c.value !== undefined && c.value !== '')
    .map(({ id: _id, ...rest }) => rest);

  if (validConditions.length === 0) return '';

  return JSON.stringify({
    logic: config.logic,
    conditions: validConditions,
  });
};

/** 布尔值选项 */
const BOOLEAN_OPTIONS = [
  { label: 'true', value: 'true' },
  { label: 'false', value: 'false' },
];

/**
 * 边条件配置表单
 */
const EdgeConditionForm: React.FC<EdgeConditionFormProps> = ({
  value,
  onChange,
  infoAtomTypes = [],
  contextKeys = [],
  contextOutputInfos = [],
}) => {
  const [config, setConfig] = useState<ConditionConfig>(() => parseConfig(value));

  // 构建信息原子字段选项（合并所有类型的字段）
  const atomFieldOptions = useMemo(() => {
    const fieldsMap = new Map<string, string>(); // fieldKey -> 显示名
    infoAtomTypes.forEach((type) => {
      type.dataFormat?.fields?.forEach((field) => {
        if (!fieldsMap.has(field.fieldKey)) {
          fieldsMap.set(field.fieldKey, `${field.fieldKey} (${field.fieldType})`);
        }
      });
    });
    return Array.from(fieldsMap.entries()).map(([key, label]) => ({
      label,
      value: key,
    }));
  }, [infoAtomTypes]);

  // 构建上下文 resultKey 选项（基于当前图中配置的 resultKey）
  const contextKeyOptions = useMemo(() => {
    if (contextKeys.length === 0) {
      return [];
    }
    return contextKeys.map((key) => ({ label: key, value: key }));
  }, [contextKeys]);

  // 构建上下文输出信息映射（contextKey -> ContextOutputInfo）
  const contextOutputMap = useMemo(() => {
    const map = new Map<string, ContextOutputInfo>();
    contextOutputInfos.forEach((info) => {
      map.set(info.contextKey, info);
    });
    return map;
  }, [contextOutputInfos]);

  // 获取指定上下文键的子字段选项
  const getSubfieldOptions = useCallback(
    (contextKey: string) => {
      const info = contextOutputMap.get(contextKey);
      if (!info?.outputSchema) return [];

      const options: { label: string; value: string }[] = [];

      // 处理 dynamic-keys 类型（如 TimeWindowCheck）
      if (info.outputSchema.type === 'dynamic-keys' && info.dynamicKeys) {
        info.dynamicKeys.forEach((key) => {
          const valueType = info.outputSchema?.valueType || 'any';
          options.push({ label: `${key} (${valueType})`, value: key });
        });
      }

      // 处理 object 类型的固定属性
      if (info.outputSchema.properties) {
        Object.entries(info.outputSchema.properties).forEach(([key, prop]) => {
          options.push({ label: `${key} (${prop.type})`, value: key });
        });
      }

      return options;
    },
    [contextOutputMap],
  );

  // 获取指定路径的值类型
  const getValueType = useCallback(
    (path: string): string | null => {
      if (!path) return null;
      const [contextKey, subfield] = path.split('.');
      if (!contextKey || !subfield) return null;

      const info = contextOutputMap.get(contextKey);
      if (!info?.outputSchema) return null;

      // dynamic-keys 类型
      if (info.outputSchema.type === 'dynamic-keys') {
        return info.outputSchema.valueType || null;
      }

      // object 类型的属性
      if (info.outputSchema.properties?.[subfield]) {
        return info.outputSchema.properties[subfield].type;
      }

      return null;
    },
    [contextOutputMap],
  );

  // 当外部 value 变化时同步
  useEffect(() => {
    setConfig(parseConfig(value));
  }, [value]);

  // 通知变更
  const notifyChange = useCallback(
    (newConfig: ConditionConfig) => {
      setConfig(newConfig);
      onChange?.(serializeConfig(newConfig));
    },
    [onChange],
  );

  // 添加条件
  const handleAddCondition = useCallback(() => {
    notifyChange({
      ...config,
      conditions: [...config.conditions, createDefaultCondition()],
    });
  }, [config, notifyChange]);

  // 删除条件
  const handleDeleteCondition = useCallback(
    (index: number) => {
      const newConditions = [...config.conditions];
      newConditions.splice(index, 1);
      notifyChange({ ...config, conditions: newConditions });
    },
    [config, notifyChange],
  );

  // 更新条件
  const handleUpdateCondition = useCallback(
    (index: number, field: keyof ConditionItem, fieldValue: string) => {
      const newConditions = [...config.conditions];
      newConditions[index] = { ...newConditions[index], [field]: fieldValue };
      notifyChange({ ...config, conditions: newConditions });
    },
    [config, notifyChange],
  );

  // 更新逻辑
  const handleLogicChange = useCallback(
    (logic: 'and' | 'or') => {
      notifyChange({ ...config, logic });
    },
    [config, notifyChange],
  );

  return (
    <div className={styles.container}>
      {/* 逻辑选择（仅在有多个条件时显示） */}
      {config.conditions.length > 1 && (
        <div className={styles.logicRow}>
          <Text className={styles.logicLabel}>条件关系：</Text>
          <ProFormSelect
            noStyle
            fieldProps={{
              value: config.logic,
              onChange: handleLogicChange,
              options: LOGIC_OPTIONS,
              size: 'small',
              className: styles.logicSelect,
            }}
          />
        </div>
      )}

      {/* 条件列表 */}
      <Space direction="vertical" size={8} className={styles.conditionList}>
        {config.conditions.map((condition, index) => (
          <Card
            key={condition.id}
            size="small"
            className={styles.conditionCard}
            extra={
              <Button
                type="text"
                danger
                size="small"
                icon={<DeleteOutlined />}
                onClick={() => handleDeleteCondition(index)}
              />
            }
          >
            <Flex vertical gap={8}>
              {/* 第一行：变量来源 */}
              <Flex align="center" gap={8}>
                <Text className={styles.fieldLabel}>变量来源</Text>
                <ProFormSelect
                  noStyle
                  fieldProps={{
                    value: condition.source,
                    onChange: (v) => handleUpdateCondition(index, 'source', v),
                    options: SOURCE_OPTIONS,
                    size: 'small',
                    className: styles.sourceSelect,
                  }}
                />
              </Flex>

              {/* 第二行：变量路径 */}
              <Flex align="center" gap={8}>
                <Text className={styles.fieldLabel}>变量路径</Text>
                {condition.source === 'atom' ? (
                  <ProFormSelect
                    noStyle
                    fieldProps={{
                      value: condition.path,
                      onChange: (v) => handleUpdateCondition(index, 'path', v),
                      options: atomFieldOptions,
                      size: 'small',
                      placeholder: '选择字段',
                      className: styles.pathInputFull,
                      showSearch: true,
                      allowClear: true,
                    }}
                  />
                ) : (
                  <Flex gap={4} align="center" className={styles.contextPathRow}>
                    <ProFormSelect
                      noStyle
                      fieldProps={{
                        value: condition.path?.split('.')[0] || undefined,
                        onChange: (v) => {
                          // 切换上下文键时清空子字段
                          handleUpdateCondition(index, 'path', v || '');
                        },
                        options: contextKeyOptions,
                        size: 'small',
                        placeholder: contextKeyOptions.length > 0 ? '上下文键' : '请先配置',
                        className: styles.contextKeySelect,
                        showSearch: true,
                        disabled: contextKeyOptions.length === 0,
                      }}
                    />
                    <Text type="secondary">.</Text>
                    {(() => {
                      const contextKey = condition.path?.split('.')[0] || '';
                      const subfieldOptions = getSubfieldOptions(contextKey);
                      // 如果有子字段选项，显示下拉选择；否则显示文本输入
                      if (subfieldOptions.length > 0) {
                        return (
                          <ProFormSelect
                            noStyle
                            fieldProps={{
                              value: condition.path?.split('.')[1] || undefined,
                              onChange: (v) => {
                                handleUpdateCondition(index, 'path', contextKey ? `${contextKey}.${v}` : v);
                              },
                              options: subfieldOptions,
                              size: 'small',
                              placeholder: '选择子字段',
                              className: styles.subfieldInput,
                              showSearch: true,
                            }}
                          />
                        );
                      }
                      return (
                        <ProFormText
                          noStyle
                          fieldProps={{
                            value: condition.path?.split('.')[1] || '',
                            onChange: (e) => {
                              const subfield = e.target.value;
                              handleUpdateCondition(index, 'path', contextKey ? `${contextKey}.${subfield}` : subfield);
                            },
                            size: 'small',
                            placeholder: '子字段',
                            className: styles.subfieldInput,
                          }}
                        />
                      );
                    })()}
                  </Flex>
                )}
              </Flex>

              {/* 第三行：操作符 + 比较值 */}
              <Flex align="center" gap={8}>
                <Text className={styles.fieldLabel}>比较条件</Text>
                <ProFormSelect
                  noStyle
                  fieldProps={{
                    value: condition.operator,
                    onChange: (v) => handleUpdateCondition(index, 'operator', v),
                    options: OPERATOR_OPTIONS,
                    size: 'small',
                    className: styles.operatorSelect,
                  }}
                />
                {(() => {
                  // 只有 context 来源才检查值类型
                  const valueType = condition.source === 'context' ? getValueType(condition.path) : null;
                  // 布尔类型显示下拉选择
                  if (valueType === 'boolean') {
                    return (
                      <ProFormSelect
                        noStyle
                        fieldProps={{
                          value: condition.value || undefined,
                          onChange: (v) => handleUpdateCondition(index, 'value', v),
                          options: BOOLEAN_OPTIONS,
                          size: 'small',
                          placeholder: '选择',
                          className: styles.valueInputFlex,
                        }}
                      />
                    );
                  }
                  // 其他类型显示文本输入
                  return (
                    <ProFormText
                      noStyle
                      fieldProps={{
                        value: condition.value,
                        onChange: (e) => handleUpdateCondition(index, 'value', e.target.value),
                        size: 'small',
                        placeholder: valueType === 'number' ? '数值' : '比较值',
                        className: styles.valueInputFlex,
                      }}
                    />
                  );
                })()}
              </Flex>
            </Flex>
          </Card>
        ))}
      </Space>

      {/* 添加条件按钮 */}
      <Button
        type="dashed"
        size="small"
        icon={<PlusOutlined />}
        onClick={handleAddCondition}
        className={styles.addBtn}
        block
      >
        添加条件
      </Button>

      {/* 空状态提示 */}
      {config.conditions.length === 0 && (
        <Text type="secondary" className={styles.emptyTip}>
          未设置条件时，边将无条件通过
        </Text>
      )}
    </div>
  );
};

export default EdgeConditionForm;
