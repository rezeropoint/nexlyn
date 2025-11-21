/**
 * AI Box 设备控制相关类型
 */

// AI Box 任务状态
export interface AIBoxTaskStatus {
  type: number; // 状态级别标识（1/2/3/4）
  style: string; // 状态级别（normal/warning/danger/success）
  label: string; // 状态描述
}

// AI Box 算法任务
export interface AIBoxAlgorithmTask {
  alg_task_session: string; // 任务会话ID
  task_desc: string; // 任务描述
  media_name: string; // 媒体名称
  alg_info: number[]; // 算法信息列表
  alg_task_status: AIBoxTaskStatus; // 任务运行状态
  alarm_body: number; // 告警主体
  alarm_protocol: number; // 告警协议
  metadata_url: string | string[]; // 元数据URL（可能是string或string[]）
  user_data?: Record<string, any>; // 用户自定义数据
}

// 算法参数选项（SELECTOR类型使用）
export interface AIBoxParameterOption {
  enable: boolean; // 当前选项是否可选
  key: string; // 选项标识
  value: number | string; // 选项序号（可能是number或空字符串）
  name: string; // 选项名称
}

// 算法可设置的配置项
export interface AIBoxParameter {
  class: string; // 设置项数据类型描述（FLOAT/BOOLEAN/SELECTOR/INTEGER）
  type: number; // 设置项数据类型编号（FLOAT=>2, BOOLEAN=>4, SELECTOR=>5, INTEGER=>0）
  max?: any; // 可设置最大值（FLOAT/INTEGER时存在）
  min?: any; // 可设置最小值（FLOAT/INTEGER时存在）
  default: any; // 设置项默认值
  key: string; // 设置项标识
  name: string; // 设置项名称
  required: boolean; // 是否必填
  value: any; // 设置值
  options?: AIBoxParameterOption[]; // SELECTOR待选列表
}

// 算法属性配置
export interface AIBoxAttribute {
  lineRequired: boolean; // 是否必须配置辅助线
  lineDesc: string; // 配置辅助线的描述文字
  zoneRequired: boolean; // 是否必须配置算法专用区域
  zoneDesc: string; // 配置算法专用区域的描述文字
}

// 算法报警规则
export interface AIBoxPolicy {
  property: string; // 违规的属性名（告警类型）
  name: string; // 属性名描述（告警类型的中文描述）
}

// 单个算法能力
export interface AIBoxAbility {
  attribute: AIBoxAttribute; // 算法属性
  parameters: AIBoxParameter[]; // 算法可设置的配置项
  permitted: boolean; // 算法是否已授权
  code: number; // 主算法编号
  sub: boolean; // 是否是子算法
  name: string; // 算法名称
  desc: string; // 算法描述
  item: number; // 子算法编号
  policy: AIBoxPolicy[]; // 报警规则
}

// AI Box 算法能力
export interface AIBoxCapabilities {
  board_id: string; // 盒子ID
  abilities: AIBoxAbility[]; // 完整算法能力列表
}
