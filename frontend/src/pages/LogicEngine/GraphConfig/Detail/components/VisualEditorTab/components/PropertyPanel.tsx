/**
 * 属性面板组件
 * @description 显示和编辑选中节点/边的属性
 */

import { getBlockSpecs, listInfoAtomTypes } from "@/services/lynxmanager/api";
import type {
  BlockSpec,
  EdgeConfig,
  InfoAtomType,
  NodeConfig,
} from "@/services/lynxmanager/types";
import { SettingOutlined } from "@ant-design/icons";
import type { FormInstance, ProFormInstance } from "@ant-design/pro-components";
import {
  ProForm,
  ProFormList,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { Card, Divider, Empty, Spin, Tag, Typography } from "antd";

const { Text } = Typography;
import React, { useEffect, useRef, useState } from "react";
import DynamicBlockConfigForm from "../../DynamicBlockConfigForm";
import styles from "./PropertyPanel.module.less";

interface PropertyPanelProps {
  selectedNode?: NodeConfig;
  selectedEdge?: EdgeConfig;
  tenantId?: string;
  onNodeUpdate?: (nodeId: string, updates: Partial<NodeConfig>) => void;
  onEdgeUpdate?: (edgeId: string, updates: Partial<EdgeConfig>) => void;
}

/**
 * 属性面板组件
 */
const PropertyPanel: React.FC<PropertyPanelProps> = ({
  selectedNode,
  selectedEdge,
  tenantId,
  onNodeUpdate,
  onEdgeUpdate,
}) => {
  const formRef = useRef<ProFormInstance>();
  const configFormRef = useRef<FormInstance>(null);
  const [blockSpecs, setBlockSpecs] = useState<Record<string, BlockSpec>>({});
  const [loadingSpecs, setLoadingSpecs] = useState(false);
  const [infoAtomTypes, setInfoAtomTypes] = useState<InfoAtomType[]>([]);
  const [loadingInfoAtomTypes, setLoadingInfoAtomTypes] = useState(false);

  // 获取所有BlockSpecs
  useEffect(() => {
    const fetchBlockSpecs = async () => {
      setLoadingSpecs(true);
      try {
        const response = await getBlockSpecs();
        if (response.code === 0 && response.data?.list) {
          const specsMap = response.data.list.reduce((acc, spec) => {
            acc[spec.blockType] = spec;
            return acc;
          }, {} as Record<string, BlockSpec>);
          setBlockSpecs(specsMap);
        }
      } catch (error) {
        console.error("Failed to fetch block specs:", error);
      } finally {
        setLoadingSpecs(false);
      }
    };

    fetchBlockSpecs();
  }, []);

  // 获取信息原子类型列表
  useEffect(() => {
    const fetchInfoAtomTypes = async () => {
      if (!tenantId) return;

      setLoadingInfoAtomTypes(true);
      try {
        const response = await listInfoAtomTypes({
          tenantId,
          page: 1,
          pageSize: 1000,
        });
        if (response.code === 0 && response.data?.list) {
          setInfoAtomTypes(response.data.list);
        }
      } catch (error) {
        console.error("Failed to fetch info atom types:", error);
      } finally {
        setLoadingInfoAtomTypes(false);
      }
    };

    fetchInfoAtomTypes();
  }, [tenantId]);

  // 当选中项变化时，更新表单值
  // 注意：只在节点/边 ID 变化时才重置表单，避免编辑时被覆盖
  useEffect(() => {
    if (selectedNode) {
      // 将 subscribedLabels ["key:value"] 转换为 ProFormList 需要的 [{key, value}] 格式
      const labelsArray = selectedNode.subscribedLabels
        ? selectedNode.subscribedLabels.map((label) => {
            const [key, ...valueParts] = label.split(":");
            return { key, value: valueParts.join(":") };
          })
        : [];

      formRef.current?.setFieldsValue({
        id: selectedNode.id,
        type: selectedNode.type,
        blockType: selectedNode.blockType,
        blockVersion: selectedNode.blockVersion,
        isEntryPoint: selectedNode.isEntryPoint,
        subscribedInfoAtomTypeIDs: selectedNode.subscribedInfoAtomTypeIDs,
        subscribedSource: selectedNode.subscribedSource || "",
        subscribedLabels: labelsArray,
      });
      configFormRef.current?.setFieldsValue(selectedNode.blockConfig || {});
    } else if (selectedEdge) {
      formRef.current?.setFieldsValue({
        id: selectedEdge.id,
        sourceID: selectedEdge.sourceID,
        targetID: selectedEdge.targetID,
        condition: selectedEdge.condition,
      });
    } else {
      formRef.current?.resetFields();
      configFormRef.current?.resetFields();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedNode?.id, selectedEdge?.id]);

  // 节点属性表单
  const renderNodeForm = () => {
    if (!selectedNode) return null;

    return (
      <ProForm
        formRef={formRef}
        layout="vertical"
        submitter={false}
        onValuesChange={(changedValues) => {
          if (!selectedNode) return;

          const updates: Partial<NodeConfig> = {};

          // 处理 subscribedInfoAtomTypeIDs：现在是数组
          if (changedValues.subscribedInfoAtomTypeIDs !== undefined) {
            updates.subscribedInfoAtomTypeIDs = Array.isArray(
              changedValues.subscribedInfoAtomTypeIDs
            )
              ? changedValues.subscribedInfoAtomTypeIDs
              : [];
          }

          // 处理 subscribedLabels：将 [{key, value}] 转换为 ["key:value"] 字符串数组
          if (changedValues.subscribedLabels !== undefined) {
            const labelsArray = changedValues.subscribedLabels || [];
            updates.subscribedLabels = labelsArray
              .filter(
                (item: { key?: string; value?: string }) => item.key && item.value
              )
              .map(
                (item: { key: string; value: string }) =>
                  `${item.key}:${item.value}`
              );
          }

          // 合并其他变化的值（subscribedSource 现在是普通字符串，无需特殊处理）
          Object.keys(changedValues).forEach((key) => {
            if (
              key !== "subscribedInfoAtomTypeIDs" &&
              key !== "subscribedLabels" &&
              key !== "id"
            ) {
              (updates as any)[key] = changedValues[key];
            }
          });

          onNodeUpdate?.(selectedNode.id, updates);
        }}
      >
        <div className={styles.formSection}>
          <div className={styles.sectionTitle}>基本信息</div>
          <ProFormText name="id" label="节点ID" disabled />
          <ProFormSelect
            name="type"
            label="节点类型"
            options={[
              { label: "入口节点", value: "entry" },
              { label: "动作节点", value: "action" },
              { label: "过滤节点", value: "filter" },
              { label: "判断节点", value: "condition" },
            ]}
          />
          <ProFormText name="blockType" label="逻辑积木类型" disabled />
          <ProFormText name="blockVersion" label="逻辑积木版本" disabled />
        </div>

        <Divider />

        <div className={styles.formSection}>
          <div className={styles.sectionTitle}>入口节点配置</div>
          <ProFormSwitch name="isEntryPoint" label="是否为入口节点" />
          {selectedNode.isEntryPoint && (
            <>
              {/* 订阅的信息原子类型ID - 下拉多选 */}
              <ProFormSelect
                name="subscribedInfoAtomTypeIDs"
                label="订阅的信息原子类型"
                mode="multiple"
                placeholder="请选择信息原子类型"
                fieldProps={{
                  showSearch: true,
                  optionFilterProp: "label",
                  loading: loadingInfoAtomTypes,
                }}
                options={infoAtomTypes.map((type) => ({
                  label: `${type.name} (v${type.version})`,
                  value: type.id,
                }))}
              />

              {/* 订阅来源 - 单一来源精确匹配 */}
              <ProFormText
                name="subscribedSource"
                label="订阅来源"
                placeholder="留空表示匹配所有来源"
                tooltip="空值 = 匹配所有来源的信息原子；填写值 = 精确匹配单一来源（如 mqtt、http、sensor001）。如需监听多个来源，请创建多个入口节点分别订阅"
              />

              {/* 订阅标签 - 键值对编辑器 */}
              <ProFormList
                name="subscribedLabels"
                label="订阅标签（键值对）"
                creatorButtonProps={{
                  position: "bottom",
                  creatorButtonText: "添加标签条件",
                  icon: false,
                }}
                copyIconProps={false}
                deleteIconProps={{ tooltipText: "删除此标签" }}
                itemRender={({ listDom, action }, { index }) => (
                  <Card
                    size="small"
                    className={styles.labelItemCard}
                    title={<Text strong>标签 #{index + 1}</Text>}
                    extra={action}
                  >
                    {listDom}
                  </Card>
                )}
              >
                <div className={styles.labelItemContainer}>
                  <ProFormText
                    name="key"
                    label="标签键"
                    placeholder="如 region、priority"
                    width="sm"
                  />
                  <ProFormText
                    name="value"
                    label="标签值"
                    placeholder="如 north、high"
                    width="sm"
                  />
                </div>
              </ProFormList>
            </>
          )}
        </div>

        <Divider />

        <div className={styles.formSection}>
          <div className={styles.sectionTitle}>逻辑积木配置</div>
          {(() => {
            try {
              const spec = selectedNode.blockType
                ? blockSpecs[selectedNode.blockType]
                : undefined;
              const configSchemaStr = spec?.configSchema;
              const configSchema = configSchemaStr
                ? JSON.parse(configSchemaStr)
                : null;

              if (configSchema && Object.keys(configSchema).length > 0) {
                return (
                  <Spin spinning={loadingSpecs}>
                    <DynamicBlockConfigForm
                      form={configFormRef}
                      schema={configSchema}
                      initialValues={
                        typeof selectedNode.blockConfig === "string"
                          ? JSON.parse(selectedNode.blockConfig || "{}")
                          : selectedNode.blockConfig || {}
                      }
                      onChange={(values) => {
                        onNodeUpdate?.(selectedNode.id, {
                          blockConfig: JSON.stringify(values),
                        });
                      }}
                    />
                  </Spin>
                );
              }
            } catch (error) {
              console.warn(
                "Failed to parse configSchema for block:",
                selectedNode.blockType,
                error
              );
            }

            return (
              <ProFormTextArea
                name="blockConfig"
                label="配置JSON"
                placeholder="输入JSON格式的配置（或选择一个有配置Schema的逻辑块类型）"
                fieldProps={{
                  rows: 8,
                  className: styles.propertyPanelJsonEditor,
                }}
              />
            );
          })()}
        </div>
      </ProForm>
    );
  };

  // 边属性表单
  const renderEdgeForm = () => {
    if (!selectedEdge) return null;

    return (
      <ProForm
        formRef={formRef}
        layout="vertical"
        submitter={false}
        onValuesChange={(changedValues) => {
          if (!selectedEdge) return;
          const updates = { ...changedValues };
          delete updates.id;
          delete updates.sourceID;
          delete updates.targetID;
          onEdgeUpdate?.(selectedEdge.id, updates);
        }}
      >
        <div className={styles.formSection}>
          <div className={styles.sectionTitle}>基本信息</div>
          <ProFormText name="id" label="边ID" disabled />
          <ProFormText name="sourceID" label="源节点ID" disabled />
          <ProFormText name="targetID" label="目标节点ID" disabled />
        </div>

        <Divider />

        <div className={styles.formSection}>
          <div className={styles.sectionTitle}>边配置</div>
          <ProFormTextArea
            name="condition"
            label="条件表达式"
            placeholder="输入条件表达式（留空表示无条件）"
            fieldProps={{
              rows: 4,
              className: styles.propertyPanelJsonEditor,
            }}
          />
        </div>
      </ProForm>
    );
  };

  // 空状态
  const renderEmpty = () => (
    <Empty
      description="请选择节点或边来查看和编辑属性"
      className={styles.propertyPanelEmpty}
    />
  );

  return (
    <Card
      title={
        <span>
          <SettingOutlined />{" "}
          {selectedNode ? (
            <>
              节点属性{" "}
              {selectedNode.isEntryPoint && (
                <Tag color="blue" className={styles.propertyPanelTag}>
                  入口
                </Tag>
              )}
            </>
          ) : selectedEdge ? (
            "边属性"
          ) : (
            "属性面板"
          )}
        </span>
      }
      size="small"
      className={styles.propertyPanelCard}
    >
      {selectedNode
        ? renderNodeForm()
        : selectedEdge
        ? renderEdgeForm()
        : renderEmpty()}
    </Card>
  );
};

export default PropertyPanel;
