/**
 * 类型定义统一导出
 */

// 公共类型
export type { BaseResponse, PageParams } from "./common";

// 设备类别和标准字段
export type { DeviceCategoryInfo, StandardFieldInfo } from "./category";

// 设备模板
export type {
  Condition,
  DataProcessingConfig,
  DeviceControlConfig,
  DispatchConfig,
  FieldCheck,
  FieldMapping,
  FilterRule,
  OnlineDetectionConfig,
  SensorTemplate,
  SensorTemplateDetail,
} from "./template";

// 设备绑定
export type {
  DeviceBinding,
  DeviceBindingSummary,
  DeviceTagSummary,
  UnboundDevice,
} from "./device";

// 设备标签
export type { DeviceTag } from "./tag";

// 平台配置
export type {
  KafkaConfig,
  PlatformDetail,
  PlatformMetadata,
  SkylarkConfig,
  WebhookConfig,
} from "./platform";

// 时序数据
export type {
  DeviceLatestValuesItem,
  DeviceStatisticsItem,
  FieldStatItem,
  FieldValue,
  TimeSeriesDataItem,
} from "./timeseries";

// AI Box 设备控制
export type {
  AIBoxAbility,
  AIBoxAlgorithmTask,
  AIBoxAttribute,
  AIBoxCapabilities,
  AIBoxParameter,
  AIBoxParameterOption,
  AIBoxPolicy,
  AIBoxTaskStatus,
} from "./control";

// HTTP 接收配置
export type {
  HttpReceiveDetail,
  HttpReceiveDispatchConfig,
  HttpReceiveFieldMapping,
  HttpReceiveMetadata,
} from "./httpReceive";
