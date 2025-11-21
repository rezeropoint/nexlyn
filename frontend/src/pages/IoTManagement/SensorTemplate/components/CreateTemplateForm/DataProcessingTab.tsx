import {
  DISPATCH_TYPE_COLORS,
  DISPATCH_TYPE_OPTIONS,
} from "@/pages/IoTManagement/PlatformManagement/constants";
import type {
  DeviceCategoryInfo,
  PlatformMetadata,
  StandardFieldInfo,
} from "@/services/iot";
import * as iotApi from "@/services/iot";
import type { InfoAtomType } from "@/services/lynxmanager";
import * as lynxmanagerApi from "@/services/lynxmanager";
import { CopyOutlined } from "@ant-design/icons";
import {
  ProFormDependency,
  ProFormDigit,
  ProFormList,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { Alert, Card, Divider, Form, Space, Spin, Tag, Typography } from "antd";
import React, { useEffect, useState } from "react";
import styles from "./DataProcessingTab.less";

const { Paragraph, Text } = Typography;

interface DataProcessingTabProps {
  deviceCategories: DeviceCategoryInfo[];
  standardFields: StandardFieldInfo[];
  fieldsLoading: boolean;
}

/**
 * 业务数据处理配置标签页组件
 */
const DataProcessingTab: React.FC<DataProcessingTabProps> = ({
  deviceCategories,
  standardFields,
  fieldsLoading,
}) => {
  // 获取form实例
  const form = Form.useFormInstance();

  // 获取当前用户信息
  const { initialState } = useModel("@@initialState");

  // 平台配置列表
  const [platforms, setPlatforms] = useState<PlatformMetadata[]>([]);
  const [platformsLoading, setPlatformsLoading] = useState(false);

  // 信息原子类型列表
  const [infoAtomTypes, setInfoAtomTypes] = useState<InfoAtomType[]>([]);
  const [infoAtomTypesLoading, setInfoAtomTypesLoading] = useState(false);

  // 字段匹配状态（key: dispatchConfigIndex-infoAtomTypeID, value: 匹配结果）
  const [fieldMatchResults, setFieldMatchResults] = useState<
    Record<string, { matched: string[]; missing: string[]; extra: string[]; isComplete: boolean }>
  >({});

  // 加载平台配置列表
  useEffect(() => {
    const fetchPlatforms = async () => {
      setPlatformsLoading(true);
      try {
        const response = await iotApi.getPlatformList({ pageSize: 100 });
        if (response.code === 0 && response.data?.list) {
          // 只返回启用的平台
          setPlatforms(response.data.list.filter((p) => p.enabled));
        }
      } catch (error) {
        console.error("获取平台配置列表失败:", error);
      } finally {
        setPlatformsLoading(false);
      }
    };

    fetchPlatforms();
  }, []);

  // 加载信息原子类型列表
  useEffect(() => {
    const fetchInfoAtomTypes = async () => {
      // 获取当前租户ID
      const tenantId = initialState?.currentUser?.tenantInfo?.tenantId;

      if (!tenantId) {
        console.warn("租户ID不存在，无法加载信息原子类型");
        return;
      }

      setInfoAtomTypesLoading(true);
      try {
        const response = await lynxmanagerApi.listInfoAtomTypes({
          tenantId,
          pageSize: 100,
        });
        if (response.code === 0 && response.data?.list) {
          setInfoAtomTypes(response.data.list);
        }
      } catch (error) {
        console.error("获取信息原子类型列表失败:", error);
      } finally {
        setInfoAtomTypesLoading(false);
      }
    };

    fetchInfoAtomTypes();
  }, [initialState?.currentUser?.tenantInfo?.tenantId]);

  /**
   * 检查字段匹配情况
   * @param infoAtomTypeID 信息原子类型ID
   * @param configuredFields 已配置的IoT标准字段列表
   */
  const checkFieldMatch = async (
    infoAtomTypeID: string,
    configuredFields: string[],
    dispatchIndex: number,
  ) => {
    try {
      // 获取信息原子类型详情
      const response = await lynxmanagerApi.getInfoAtomType(infoAtomTypeID);
      if (response.code !== 0 || !response.data) {
        return;
      }

      const infoAtomType = response.data;
      const requiredFields = infoAtomType.dataFormat.fields.map((f) => f.fieldKey);

      // 计算匹配结果
      const matched = configuredFields.filter((f) => requiredFields.includes(f));
      const missing = requiredFields.filter((f) => !configuredFields.includes(f));
      const extra = configuredFields.filter((f) => !requiredFields.includes(f));
      const isComplete = missing.length === 0;

      // 更新匹配状态
      const matchKey = `${dispatchIndex}-${infoAtomTypeID}`;
      setFieldMatchResults((prev) => ({
        ...prev,
        [matchKey]: { matched, missing, extra, isComplete },
      }));
    } catch (error) {
      console.error("检查字段匹配失败:", error);
    }
  };

  return (
    <>
      <Alert
        message="业务数据处理配置"
        description="处理设备上报的业务数据，包括数据提取、字段映射、数据验证等。不配置则设备数据不会被解析和存储。"
        type="info"
        showIcon
        className={styles.alertContainer}
      />

      <ProFormText
        name="businessTopicSuffix"
        label="MQTT主题后缀"
        placeholder="例如：data、telemetry、sensor"
        tooltip="用于业务数据的MQTT主题后缀，将与类别和设备ID组合成完整主题"
        rules={[
          {
            pattern: /^[a-zA-Z0-9_/-]+$/,
            message: "只能包含字母、数字、下划线、中划线和斜杠",
          },
        ]}
      />

      {/* 完整主题预览 */}
      <ProFormDependency name={["category", "model", "businessTopicSuffix"]}>
        {({ category, model, businessTopicSuffix }) => {
          if (category && model && businessTopicSuffix) {
            const fullTopic = `nexlyn/iot/${category}/${model}/{device_id}/${businessTopicSuffix}`;
            return (
              <Card size="small" className={styles.topicPreviewCard}>
                <Space
                  direction="vertical"
                  size={4}
                  className={styles.previewContent}
                >
                  <Text type="secondary" className={styles.previewLabel}>
                    完整MQTT主题（可复制）：
                  </Text>
                  <Paragraph
                    copyable={{
                      text: fullTopic,
                      icon: [
                        <CopyOutlined key="copy-icon" />,
                        <Text key="copy-success" type="success">
                          已复制
                        </Text>,
                      ],
                    }}
                    className={styles.previewTopic}
                  >
                    {fullTopic}
                  </Paragraph>
                  <Text type="secondary" className={styles.previewHint}>
                    提示：{"{device_id}"} 将被替换为实际的设备ID
                  </Text>
                </Space>
              </Card>
            );
          }
          return null;
        }}
      </ProFormDependency>

      <Divider orientation="left" className={styles.dividerSubsection}>
        时间戳处理
      </Divider>

      <Alert
        message="时间戳字段说明"
        description="配置如何从MQTT消息中提取时间戳。如果不配置或字段不存在，系统将使用服务器接收消息的时间。"
        type="info"
        showIcon
        className={styles.alertInfo}
      />

      <ProFormText
        name="timestampPath"
        label="时间戳路径"
        placeholder="例如：timestamp、data.time、ts"
        tooltip="时间戳在MQTT消息JSON中的字段路径，留空则使用服务器时间"
      />

      <ProFormSelect
        name="timestampFormat"
        label="时间戳格式"
        tooltip="消息中时间戳的格式"
        options={[
          { label: "Unix时间戳（秒）", value: "unix" },
          { label: "Unix时间戳（毫秒）", value: "unix_ms" },
          { label: "ISO 8601", value: "iso8601" },
          { label: "RFC 3339", value: "rfc3339" },
        ]}
      />

      <Divider orientation="left" className={styles.dividerSection}>
        字段映射
      </Divider>

      <ProFormDependency name={["category"]}>
        {({ category }) => {
          const categoryInfo = deviceCategories.find(
            (c) => c.code === category
          );
          const requiredFields = standardFields.filter((f) => f.required);

          if (!category) {
            return (
              <Alert
                message="请先选择设备类别"
                description="在基础信息标签页中选择设备类别后，系统将自动加载该类别的标准字段定义"
                type="warning"
                showIcon
                className={styles.warningAlert}
              />
            );
          }

          return (
            <Card size="small" className={styles.categoryInfoCard}>
              <Space
                direction="vertical"
                size={4}
                className={styles.categoryContent}
              >
                <Text strong className={styles.categoryName}>
                  当前类别：{categoryInfo?.name || category}
                </Text>
                {categoryInfo?.description && (
                  <Text type="secondary" className={styles.categoryDescription}>
                    {categoryInfo.description}
                  </Text>
                )}
                {requiredFields.length > 0 && (
                  <div className={styles.requiredFieldsRow}>
                    <Text type="secondary" className={styles.requiredLabel}>
                      必填字段：
                    </Text>
                    <Text strong className={styles.requiredFields}>
                      {requiredFields.map((f) => f.displayName).join("、")}
                    </Text>
                  </div>
                )}
                <Text type="secondary" className={styles.availableFieldsHint}>
                  可用标准字段数：{standardFields.length} 个
                </Text>
              </Space>
            </Card>
          );
        }}
      </ProFormDependency>

      <Spin spinning={fieldsLoading}>
        <ProFormList
          name="fieldMappings"
          label="字段映射列表"
          tooltip="定义如何从MQTT消息中提取数据并映射到平台标准字段"
          creatorButtonProps={{
            position: "bottom",
            creatorButtonText: "+ 添加字段映射",
          }}
          copyIconProps={false}
          deleteIconProps={{ tooltipText: "删除此映射" }}
          itemRender={({ listDom, action }, { index }) => (
            <Card
              size="small"
              className={styles.fieldMappingCard}
              title={
                <Space>
                  <Text strong className={styles.cardTitle}>
                    映射规则 #{index + 1}
                  </Text>
                </Space>
              }
              extra={action}
            >
              {listDom}
            </Card>
          )}
        >
          <Space
            direction="vertical"
            className={styles.mappingContent}
            size="middle"
          >
            <ProFormSelect
              name="standardField"
              label="平台标准字段"
              placeholder={
                standardFields.length > 0
                  ? "请选择标准字段"
                  : "请先在基础信息中选择设备类别"
              }
              disabled={standardFields.length === 0}
              showSearch
              options={standardFields.map((field) => {
                const label = field.unit
                  ? `${field.displayName} (${field.name}) - ${field.unit}`
                  : `${field.displayName} (${field.name})`;
                const description = [];
                if (field.required) description.push("必填");
                if (
                  field.minValue !== undefined ||
                  field.maxValue !== undefined
                ) {
                  const range = `${field.minValue ?? "-∞"} ~ ${
                    field.maxValue ?? "+∞"
                  }`;
                  description.push(`范围: ${range}`);
                }
                if (field.enumValues && field.enumValues.length > 0) {
                  description.push(`枚举: ${field.enumValues.join("/")}`);
                }

                return {
                  label,
                  value: field.name,
                  description: description.join(" | "),
                  fieldInfo: field,
                };
              })}
              fieldProps={{
                optionFilterProp: "label",
                optionLabelProp: "label",
              }}
              tooltip="选择平台预定义的标准字段。标准字段包含类型、单位、取值范围等元数据"
              rules={[{ required: true, message: "请选择标准字段" }]}
            />

            <ProFormText
              name="sourcePath"
              label="源字段路径"
              placeholder="例如：data.temp、sensors.temperature、temp"
              tooltip="MQTT消息中的数据字段路径，使用点分隔符"
              rules={[{ required: true, message: "请输入源字段路径" }]}
            />

            <ProFormSelect
              name="fieldType"
              label="字段类型"
              placeholder="请选择字段类型"
              options={[
                { label: "字符串", value: "string" },
                { label: "整数", value: "int" },
                { label: "浮点数", value: "float" },
                { label: "布尔值", value: "boolean" },
                { label: "时间戳", value: "timestamp" },
                { label: "JSON对象", value: "object" },
              ]}
              tooltip="选择标准字段后会自动设置，也可以手动修改"
            />

            <ProFormDigit
              name="scale"
              label="缩放因子"
              placeholder="1"
              tooltip="数值乘以此因子，例如温度单位转换（℃转K需乘1）"
              fieldProps={{ precision: 4 }}
            />

            <ProFormDigit
              name="offset"
              label="偏移量"
              placeholder="0"
              tooltip="缩放后加上此偏移量，例如温度单位转换（℃转K需加273.15）"
              fieldProps={{ precision: 4 }}
            />

            <ProFormText
              name="defaultValue"
              label="默认值"
              placeholder="当源字段不存在时使用的默认值"
              tooltip="可选，当源字段不存在或为空时使用此值"
            />
          </Space>
        </ProFormList>
      </Spin>

      <Divider orientation="left" className={styles.dividerSection}>
        数据过滤（可选）
      </Divider>

      <Alert
        message="数据过滤规则说明"
        description="配置数据过滤规则，用于在ClickHouse存储后、第三方平台分发前筛选数据。只有满足过滤条件的数据才会被分发。"
        type="info"
        showIcon
        className={styles.alertInfo}
      />

      <ProFormDependency name={["fieldMappings"]}>
        {({ fieldMappings }) => {
          const fieldOptions = (fieldMappings || []).map((fm: any) => ({
            label: fm.standardField,
            value: fm.standardField,
          }));

          return (
            <>
              <ProFormSelect
                name={["filterRules", "logic"]}
                label="条件逻辑关系"
                placeholder="请选择条件间的逻辑关系"
                options={[
                  { label: "AND（所有条件都满足）", value: "AND" },
                  { label: "OR（任一条件满足）", value: "OR" },
                ]}
                tooltip="多个条件之间的逻辑关系：AND表示所有条件都必须满足，OR表示满足任一条件即可"
              />

              <ProFormList
                name={["filterRules", "conditions"]}
                label="过滤条件列表"
                tooltip="定义具体的过滤条件，支持数值比较、字符串包含、枚举值判断等"
                creatorButtonProps={{
                  position: "bottom",
                  creatorButtonText: "+ 添加过滤条件",
                }}
                copyIconProps={false}
                deleteIconProps={{ tooltipText: "删除此条件" }}
                itemRender={({ listDom, action }, { index }) => (
                  <Card
                    size="small"
                    className={styles.filterConditionCard}
                    title={
                      <Space>
                        <Text strong className={styles.cardTitle}>
                          条件 #{index + 1}
                        </Text>
                      </Space>
                    }
                    extra={action}
                  >
                    {listDom}
                  </Card>
                )}
              >
                <Space
                  direction="vertical"
                  className={styles.conditionContent}
                  size="middle"
                >
                  <ProFormSelect
                    name="field"
                    label="字段名"
                    placeholder={
                      fieldOptions.length > 0
                        ? "请选择字段"
                        : "请先配置字段映射"
                    }
                    disabled={fieldOptions.length === 0}
                    options={fieldOptions}
                    tooltip="选择要进行过滤判断的字段（来自字段映射中的标准字段）"
                    rules={[{ required: true, message: "请选择字段" }]}
                  />

                  <ProFormSelect
                    name="operator"
                    label="操作符"
                    placeholder="请选择比较操作符"
                    rules={[{ required: true, message: "请选择操作符" }]}
                    options={[
                      { label: "等于 (=)", value: "eq" },
                      { label: "不等于 (≠)", value: "ne" },
                      { label: "大于 (>)", value: "gt" },
                      { label: "小于 (<)", value: "lt" },
                      { label: "大于等于 (≥)", value: "gte" },
                      { label: "小于等于 (≤)", value: "lte" },
                      { label: "包含", value: "contains" },
                      { label: "在列表中", value: "in" },
                      { label: "不在列表中", value: "not_in" },
                    ]}
                    tooltip="选择比较操作符"
                  />

                  <ProFormDependency name={["operator"]}>
                    {({ operator }) => {
                      if (operator === "in" || operator === "not_in") {
                        return (
                          <ProFormText
                            name="value"
                            label="比较值（数组）"
                            placeholder='例如：["value1", "value2", "value3"]'
                            tooltip="输入JSON数组格式的值列表，用于in/not_in操作"
                            rules={[
                              { required: true, message: "请输入比较值" },
                            ]}
                          />
                        );
                      }
                      return (
                        <ProFormText
                          name="value"
                          label="比较值"
                          placeholder="输入要比较的值"
                          tooltip="输入用于比较的值"
                          rules={[{ required: true, message: "请输入比较值" }]}
                        />
                      );
                    }}
                  </ProFormDependency>
                </Space>
              </ProFormList>
            </>
          );
        }}
      </ProFormDependency>

      <Divider orientation="left" className={styles.dividerSection}>
        数据分发（可选）
      </Divider>

      <Alert
        message="数据分发配置说明"
        description="配置数据的分发目标，支持多种分发类型。处理后的数据将按配置分发到对应的平台或存储。"
        type="info"
        showIcon
        className={styles.alertInfo}
      />

      <Spin spinning={platformsLoading}>
        <ProFormList
          name="dispatchConfigs"
          label="分发配置列表"
          tooltip="定义数据的分发规则，支持多种分发目标"
          creatorButtonProps={{
            position: "bottom",
            creatorButtonText: "+ 添加分发配置",
          }}
          copyIconProps={false}
          deleteIconProps={{ tooltipText: "删除此配置" }}
          itemRender={({ listDom, action }, { index, record: _record }) => {
            // 将index存储到隐藏字段中，供内部组件使用
            return (
              <Card
                size="small"
                className={styles.dispatchConfigCard}
                title={
                  <Space>
                    <Text strong className={styles.cardTitle}>
                      分发配置 #{index + 1}
                    </Text>
                  </Space>
                }
                extra={action}
                data-dispatch-index={index}
              >
                {listDom}
              </Card>
            );
          }}
        >
          <Space
            direction="vertical"
            className={styles.configContent}
            size="middle"
          >
            <ProFormSelect
              name="type"
              label="分发类型"
              placeholder="请选择分发类型"
              rules={[{ required: true, message: "请选择分发类型" }]}
              options={DISPATCH_TYPE_OPTIONS.map((opt) => ({
                label: opt.label,
                value: opt.value,
              }))}
              tooltip="选择数据的分发目标类型"
              fieldProps={{
                optionRender: (option) => (
                  <Space>
                    <Tag color={DISPATCH_TYPE_COLORS[option.value as string]}>
                      {option.label}
                    </Tag>
                  </Space>
                ),
              }}
            />

            {/* 根据分发类型动态显示字段 */}
            <ProFormDependency name={[["type"], ["fieldMappings"], ["infoAtomTypeID"]]}>
              {({ type, fieldMappings, infoAtomTypeID }) => {
                // Skylark流程触发
                if (type === "skylark_flows") {
                  return (
                    <>
                      <ProFormSelect
                        name="platformID"
                        label="Skylark平台"
                        placeholder="请选择Skylark平台配置"
                        tooltip="选择要使用的Skylark平台配置"
                        rules={[
                          { required: true, message: "请选择Skylark平台" },
                        ]}
                        options={platforms
                          .filter((p) => p.type === "skylark")
                          .map((p) => ({
                            label: `${p.name} (${p.id})`,
                            value: p.id,
                          }))}
                        fieldProps={{
                          showSearch: true,
                          optionFilterProp: "label",
                        }}
                      />

                      <ProFormDigit
                        name="flowID"
                        label="流程ID"
                        placeholder="输入Skylark流程ID"
                        tooltip="要触发的Skylark流程ID"
                        rules={[{ required: true, message: "请输入流程ID" }]}
                        fieldProps={{ precision: 0, min: 1 }}
                      />
                    </>
                  );
                }

                // Skylark表单提交
                if (type === "skylark_forms") {
                  return (
                    <>
                      <ProFormSelect
                        name="platformID"
                        label="Skylark平台"
                        placeholder="请选择Skylark平台配置"
                        tooltip="选择要使用的Skylark平台配置"
                        rules={[
                          { required: true, message: "请选择Skylark平台" },
                        ]}
                        options={platforms
                          .filter((p) => p.type === "skylark")
                          .map((p) => ({
                            label: `${p.name} (${p.id})`,
                            value: p.id,
                          }))}
                        fieldProps={{
                          showSearch: true,
                          optionFilterProp: "label",
                        }}
                      />

                      <ProFormDigit
                        name="formID"
                        label="表单ID"
                        placeholder="输入Skylark表单ID"
                        tooltip="要提交的Skylark表单ID"
                        rules={[{ required: true, message: "请输入表单ID" }]}
                        fieldProps={{ precision: 0, min: 1 }}
                      />
                    </>
                  );
                }

                // LynxGraph逻辑引擎
                if (type === "lynxgraph") {
                  // 获取已配置的字段列表
                  const configuredFields = (fieldMappings || []).map(
                    (m: any) => m.standardField,
                  );

                  // 获取当前dispatch配置的索引（使用时间戳作为唯一标识）
                  const dispatchConfigs = form.getFieldValue("dispatchConfigs") || [];
                  const currentIndex =
                    dispatchConfigs.findIndex((dc: any) => dc === ({ type, infoAtomTypeID })) ||
                    dispatchConfigs.length - 1;

                  //生成匹配键
                  const matchKey = infoAtomTypeID ? `${currentIndex}-${infoAtomTypeID}` : "";
                  const result = matchKey ? fieldMatchResults[matchKey] : null;

                  return (
                    <>
                      <Alert
                        message="字段验证说明"
                        description="请先在上方【字段映射配置】中配置要提取的标准字段，然后选择信息原子类型。系统会自动检查已配置的字段是否与信息原子类型的字段定义完全一致，只有完全一致时才能保存配置。"
                        type="info"
                        showIcon
                        style={{ marginBottom: 16 }}
                      />

                      <ProFormSelect
                        name="infoAtomTypeID"
                        label="信息原子类型"
                        placeholder="请选择信息原子类型"
                        tooltip="选择后系统会检查字段映射配置的标准字段是否与信息原子类型的字段定义完全一致"
                        rules={[
                          {
                            required: true,
                            message: "请选择信息原子类型",
                          },
                          {
                            validator: async (_, value) => {
                              if (!value) return Promise.resolve();

                              // 检查是否配置了字段映射
                              if (configuredFields.length === 0) {
                                return Promise.reject(
                                  new Error("请先在上方【字段映射配置】中配置要提取的标准字段"),
                                );
                              }

                              // 检查字段是否完全匹配
                              if (result && !result.isComplete) {
                                return Promise.reject(
                                  new Error(`字段不完全匹配：缺少 ${result.missing.join(", ")}`),
                                );
                              }

                              // 如果没有匹配结果，需要先触发检查
                              if (!result) {
                                return Promise.reject(
                                  new Error("正在验证字段匹配，请稍候..."),
                                );
                              }

                              return Promise.resolve();
                            },
                          },
                        ]}
                        options={infoAtomTypes.map((atomType) => ({
                          label: `${atomType.name} (${atomType.id})`,
                          value: atomType.id,
                        }))}
                        fieldProps={{
                          showSearch: true,
                          optionFilterProp: "label",
                          loading: infoAtomTypesLoading,
                          onChange: (value: string) => {
                            if (value) {
                              // 无论是否有字段映射都要触发检查，这样验证器才能正确提示错误
                              checkFieldMatch(value, configuredFields, currentIndex);
                            }
                          },
                        }}
                      />

                      {/* 字段匹配状态显示 */}
                      {infoAtomTypeID && (
                        <Card size="small" style={{ marginTop: 16 }}>
                          <Text strong>字段匹配状态</Text>
                          <div style={{ marginTop: 12 }}>
                            {!result && configuredFields.length === 0 && (
                              <Alert
                                message="未配置字段映射"
                                description="请先在上方【字段映射配置】中配置要提取的标准字段，然后系统会自动检查是否与所选信息原子类型匹配。"
                                type="warning"
                                showIcon
                              />
                            )}
                            {result && result.matched.length > 0 && (
                              <div style={{ marginBottom: 8 }}>
                                <Tag color="success">已匹配 ({result.matched.length})</Tag>
                                {result.matched.map((field) => (
                                  <Tag key={field}>{field}</Tag>
                                ))}
                              </div>
                            )}
                            {result && result.missing.length > 0 && (
                              <div style={{ marginBottom: 8 }}>
                                <Tag color="error">缺失 ({result.missing.length})</Tag>
                                {result.missing.map((field) => (
                                  <Tag key={field} color="red">
                                    {field}
                                  </Tag>
                                ))}
                              </div>
                            )}
                            {result && result.extra.length > 0 && (
                              <div>
                                <Tag color="warning">未使用 ({result.extra.length})</Tag>
                                {result.extra.map((field) => (
                                  <Tag key={field} color="orange">
                                    {field}
                                  </Tag>
                                ))}
                              </div>
                            )}
                            {result && (
                              result.isComplete ? (
                                <Alert
                                  message="字段完全匹配，可以保存"
                                  type="success"
                                  showIcon
                                  style={{ marginTop: 12 }}
                                />
                              ) : (
                                <Alert
                                  message="字段不完全匹配，请在上方【字段映射配置】中添加或移除字段，确保与信息原子类型完全一致"
                                  type="error"
                                  showIcon
                                  style={{ marginTop: 12 }}
                                />
                              )
                            )}
                          </div>
                        </Card>
                      )}
                    </>
                  );
                }

                // 日志记录不需要额外配置
                if (type === "log") {
                  return (
                    <Alert
                      message="日志记录"
                      description="数据将记录到系统日志，包含提取的字段和原始消息。"
                      type="info"
                      showIcon
                    />
                  );
                }

                return null;
              }}
            </ProFormDependency>

            <ProFormTextArea
              name="extraParams"
              label="扩展参数"
              placeholder="请输入JSON格式的扩展参数（可选）"
              tooltip="JSON格式的额外参数，用于特定平台的自定义配置"
              fieldProps={{
                rows: 3,
                className: styles.extraParamsTextarea,
              }}
            />
          </Space>
        </ProFormList>
      </Spin>
    </>
  );
};

export default DataProcessingTab;
