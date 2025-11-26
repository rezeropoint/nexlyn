import type { PlatformMetadata } from "@/services/iot";
import * as iotApi from "@/services/iot";
import type { InfoAtomType } from "@/services/lynxmanager";
import * as lynxmanagerApi from "@/services/lynxmanager";
import {
  ProFormDependency,
  ProFormDigit,
  ProFormList,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { Alert, Card, Divider, Space, Spin, Tag, Typography } from "antd";
import React, { useEffect, useState } from "react";
import {
  DISPATCH_TYPE_COLORS,
  DISPATCH_TYPE_OPTIONS,
  FIELD_TYPE_OPTIONS,
  TIMESTAMP_FORMAT_OPTIONS,
} from "../../constants";

const { Text } = Typography;

/**
 * 数据处理配置标签页组件
 *
 * HTTP 接收功能处理任意格式的 JSON 数据，使用自定义字段名（不再映射到标准字段）
 * 包含：时间戳配置、设备ID配置、字段映射、分发配置
 */
const DataProcessingTab: React.FC = () => {
  // 获取当前用户信息
  const { initialState } = useModel("@@initialState");

  // 平台配置列表
  const [platforms, setPlatforms] = useState<PlatformMetadata[]>([]);
  const [platformsLoading, setPlatformsLoading] = useState(false);

  // 信息原子类型列表
  const [infoAtomTypes, setInfoAtomTypes] = useState<InfoAtomType[]>([]);
  const [infoAtomTypesLoading, setInfoAtomTypesLoading] = useState(false);

  // 加载平台配置列表
  useEffect(() => {
    const fetchPlatforms = async () => {
      setPlatformsLoading(true);
      try {
        const response = await iotApi.getPlatformList({ pageSize: 100 });
        if (response.code === 0 && response.data?.list) {
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
      const tenantId = initialState?.currentUser?.tenantInfo?.tenantId;
      if (!tenantId) return;

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

  return (
    <>
      <Alert
        message="数据处理与分发配置"
        description="配置如何从 HTTP 请求体中提取数据，以及数据的分发目标。请求体必须是 JSON 格式。字段名可自定义，无需映射到预定义的标准字段。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      {/* ===== 时间戳配置 ===== */}
      <Divider orientation="left">时间戳配置</Divider>

      <ProFormText
        name="timestampPath"
        label="时间戳字段路径"
        placeholder="例如：timestamp、data.time、ts"
        tooltip="时间戳在请求体 JSON 中的字段路径，留空则使用服务器时间"
      />

      <ProFormSelect
        name="timestampFormat"
        label="时间戳格式"
        placeholder="请选择时间戳格式"
        options={TIMESTAMP_FORMAT_OPTIONS}
        tooltip="请求体中时间戳的格式"
      />

      {/* ===== 设备ID配置 ===== */}
      <Divider orientation="left">设备ID配置</Divider>

      <ProFormText
        name="deviceIdPath"
        label="设备ID字段路径"
        placeholder="例如：deviceId、data.device_id、id"
        tooltip="设备ID在请求体 JSON 中的字段路径，用于标识数据来源设备"
      />

      {/* ===== 字段映射配置 ===== */}
      <Divider orientation="left">字段映射配置</Divider>

      <Alert
        message="自定义字段映射"
        description="定义要从 JSON 请求体中提取的字段。字段名由您自定义，可根据业务需要命名（如 temperature、humidity 等）。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <ProFormList
        name="fieldMappings"
        label="字段映射列表"
        tooltip="定义如何从请求体中提取数据字段"
        creatorButtonProps={{
          position: "bottom",
          creatorButtonText: "+ 添加字段映射",
        }}
        copyIconProps={false}
        deleteIconProps={{ tooltipText: "删除此映射" }}
        itemRender={({ listDom, action }, { index }) => (
          <Card
            size="small"
            style={{ marginBottom: 8 }}
            title={
              <Text strong style={{ fontSize: 13 }}>
                映射规则 #{index + 1}
              </Text>
            }
            extra={action}
          >
            {listDom}
          </Card>
        )}
      >
        <Space direction="vertical" style={{ width: "100%" }} size="middle">
          <ProFormText
            name="fieldName"
            label="字段名称"
            placeholder="例如：temperature、humidity、status"
            tooltip="自定义的字段名称，将作为分发数据中的 key"
            rules={[{ required: true, message: "请输入字段名称" }]}
          />

          <ProFormText
            name="sourcePath"
            label="源字段路径"
            placeholder="例如：data.temp、sensors.temperature、temp"
            tooltip="请求体 JSON 中的数据字段路径，使用点分隔符"
            rules={[{ required: true, message: "请输入源字段路径" }]}
          />

          <ProFormSelect
            name="fieldType"
            label="字段类型"
            placeholder="请选择字段类型"
            options={FIELD_TYPE_OPTIONS}
            tooltip="数据字段的类型"
          />

          <ProFormText
            name="defaultValue"
            label="默认值"
            placeholder="当源字段不存在时使用的默认值"
            tooltip="可选，当源字段不存在或为空时使用此值"
          />
        </Space>
      </ProFormList>

      {/* ===== 分发配置 ===== */}
      <Divider orientation="left">分发配置</Divider>

      <Spin spinning={platformsLoading}>
        <ProFormList
          name="dispatchConfigs"
          label="分发配置列表"
          tooltip="定义数据的分发规则，支持多种分发目标，数据可同时分发到多个平台"
          creatorButtonProps={{
            position: "bottom",
            creatorButtonText: "+ 添加分发配置",
          }}
          copyIconProps={false}
          deleteIconProps={{ tooltipText: "删除此配置" }}
          itemRender={({ listDom, action }, { index }) => (
            <Card
              size="small"
              style={{ marginBottom: 8 }}
              title={
                <Text strong style={{ fontSize: 13 }}>
                  分发配置 #{index + 1}
                </Text>
              }
              extra={action}
            >
              {listDom}
            </Card>
          )}
        >
          <Space direction="vertical" style={{ width: "100%" }} size="middle">
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
            <ProFormDependency name={["type"]}>
              {({ type }) => {
                // Skylark 流程触发
                if (type === "skylark_flows") {
                  return (
                    <>
                      <ProFormSelect
                        name="platformId"
                        label="Skylark 平台"
                        placeholder="请选择 Skylark 平台配置"
                        tooltip="选择要使用的 Skylark 平台配置"
                        rules={[
                          { required: true, message: "请选择 Skylark 平台" },
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
                        name="flowId"
                        label="流程ID"
                        placeholder="输入 Skylark 流程ID"
                        tooltip="要触发的 Skylark 流程ID"
                        rules={[{ required: true, message: "请输入流程ID" }]}
                        fieldProps={{ precision: 0, min: 1 }}
                      />
                    </>
                  );
                }

                // Skylark 表单提交
                if (type === "skylark_forms") {
                  return (
                    <>
                      <ProFormSelect
                        name="platformId"
                        label="Skylark 平台"
                        placeholder="请选择 Skylark 平台配置"
                        tooltip="选择要使用的 Skylark 平台配置"
                        rules={[
                          { required: true, message: "请选择 Skylark 平台" },
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
                        name="formId"
                        label="表单ID"
                        placeholder="输入 Skylark 表单ID"
                        tooltip="要提交的 Skylark 表单ID"
                        rules={[{ required: true, message: "请输入表单ID" }]}
                        fieldProps={{ precision: 0, min: 1 }}
                      />
                    </>
                  );
                }

                // LynxGraph 逻辑引擎
                if (type === "lynxgraph") {
                  return (
                    <ProFormSelect
                      name="infoAtomTypeId"
                      label="信息原子类型"
                      placeholder="请选择信息原子类型"
                      tooltip="选择要使用的 LynxGraph 信息原子类型"
                      rules={[
                        { required: true, message: "请选择信息原子类型" },
                      ]}
                      options={infoAtomTypes.map((atomType) => ({
                        label: `${atomType.name} (${atomType.id})`,
                        value: atomType.id,
                      }))}
                      fieldProps={{
                        showSearch: true,
                        optionFilterProp: "label",
                        loading: infoAtomTypesLoading,
                      }}
                    />
                  );
                }

                // 日志记录
                if (type === "log") {
                  return (
                    <Alert
                      message="日志记录"
                      description="数据将记录到系统日志，用于调试和监控。"
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
              placeholder='请输入 JSON 格式的扩展参数（可选），例如：{"key": "value"}'
              tooltip="JSON 格式的额外参数，用于特定平台的自定义配置"
              fieldProps={{
                rows: 2,
              }}
            />
          </Space>
        </ProFormList>
      </Spin>
    </>
  );
};

export default DataProcessingTab;
