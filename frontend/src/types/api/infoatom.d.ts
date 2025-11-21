// 信息原子管理相关类型定义

declare namespace API {
  /** 字段类型 */
  type FieldType = {
    value: string; // 'string' | 'int' | 'float' | 'bool'
  };

  /** 字段配置 */
  type FieldConfig = {
    fieldKey: string;
    fieldPath: string;
    fieldType: FieldType;
  };

  /** 数据格式 */
  type DataFormat = {
    dataPlural: boolean;
    fieldStart?: string;
    fields: FieldConfig[];
  };

  /** 信息原子类型 */
  type InfoAtomType = {
    id: string;
    tenantId: string;
    name: string;
    version: string;
    tags: string[];
    dataFormat: DataFormat;
    createdAt?: number;
    updatedAt?: number;
  };

  /** 列表查询请求 */
  type GetInfoAtomTypeListRequest = PageParams & {
    tenantId?: string;
    name?: string;
    tags?: string[];
  };

  /** 列表数据 */
  type InfoAtomTypeListData = {
    list: InfoAtomType[];
  };

  /** 列表响应 */
  type GetInfoAtomTypeListResponse = BaseResponse &
    PageParams & {
      data?: InfoAtomTypeListData;
    };

  /** 创建请求 */
  type CreateInfoAtomTypeRequest = {
    tenantId: string;
    name: string;
    version: string;
    tags: string[];
    dataFormat: DataFormat;
  };

  /** 创建响应 */
  type CreateInfoAtomTypeResponse = BaseResponse & {
    data?: { id: string };
  };

  /** 更新请求 */
  type UpdateInfoAtomTypeRequest = {
    tenantId?: string;
    name?: string;
    version?: string;
    tags?: string[];
    dataFormat?: DataFormat;
  };

  /** 更新响应 */
  type UpdateInfoAtomTypeResponse = BaseResponse;

  /** 删除响应 */
  type DeleteInfoAtomTypeResponse = BaseResponse;

  /** 获取详情响应 */
  type GetInfoAtomTypeResponse = BaseResponse & {
    data?: InfoAtomType;
  };

  /** 检查存在响应 */
  type CheckInfoAtomTypeResponse = BaseResponse & {
    data?: { exists: boolean };
  };
}
