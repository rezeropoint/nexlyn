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
import { Card, Divider, Empty, Space, Spin, Tag } from "antd";
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
  useEffect(() => {
    if (selectedNode) {
      // 将 subscribedLabels map 转换为 ProFormList 需要的数组格式
      const labelsArray = selectedNode.subscribedLabels
        ? Object.entries(selectedNode.subscribedLabels).map(([key, value]) => ({
            key,
            value,
          }))
        : [];

      formRef.current?.setFieldsValue({
        id: selectedNode.id,
        type: selectedNode.type,
        blockType: selectedNode.blockType,
        blockVersion: selectedNode.blockVersion,
        isEntryPoint: selectedNode.isEntryPoint,
        subscribedInfoAtomTypeIDs: selectedNode.subscribedInfoAtomTypeIDs,
        subscribedSource: selectedNode.subscribedSource,
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
  }, [selectedNode, selectedEdge]);

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

          // 处理 subscribedLabels：现在是 [{key, value}] 数组，需要转换为 map[string]string
          if (changedValues.subscribedLabels !== undefined) {
            const labelsArray = changedValues.subscribedLabels || [];
            updates.subscribedLabels = labelsArray.reduce(
              (
                acc: Record<string, string>,
                item: { key?: string; value?: string }
              ) => {
                if (item.key && item.value) {
                  acc[item.key] = item.value;
                }
                return acc;
              },
              {}
            );
          }

          // 合并其他变化的值
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

              {/* 订阅来源 - 自定义输入 */}
              <ProFormSelect
                name="subscribedSource"
                label="订阅来源"
                mode="tags"
                placeholder="输入设备ID（如 sensor001）"
                tooltip="留空表示接收任何设备的数据；填入设备ID则只接收特定设备的数据"
                fieldProps={{
                  maxTagCount: 1,
                  maxTagTextLength: 100,
                  allowClear: true,
                }}
                options={[]}
                extra={
                  <div className={styles.fieldTip}>
                    <span className={styles.tipLabel}>示例：</span>
                    <span className={styles.tipContent}>
                      sensor001、device_001、box_123
                    </span>
                  </div>
                }
              />

              {/* 订阅标签 - 键值对编辑器 */}
              <ProFormList
                name="subscribedLabels"
                label="订阅标签（键值对）"
                creatorButtonProps={{
                  creatorButtonText: "添加标签条件",
                }}
                itemRender={({ listDom, action }, { index }) => (
                  <Space key={index} style={{ width: "100%" }} align="baseline">
                    {listDom}
                    {action}
                  </Space>
                )}
              >
                <ProFormText
                  name="key"
                  label="标签键"
                  placeholder="如 region、priority"
                  width="xs"
                />
                <ProFormText
                  name="value"
                  label="标签值"
                  placeholder="如 north、high"
                  width="xs"
                />
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
