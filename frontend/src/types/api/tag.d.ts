// 标签管理相关类型定义

declare namespace API {
  /** 标签定义（区分用户与租户） */
  type TagDefinition = {
    id: string;
    scope: string; // tenant | user
    label: string; // 标签名称
    description: string; // 描述
    createdAt: number;
    updatedAt: number;
  };

  /** 标签简项（用于下拉等） */
  type TagOption = {
    id: string;
    label: string;
    count: number;
  };

  /** 创建标签请求 */
  type CreateTagRequest = {
    scope: string; // tenant | user（必填）
    label: string; // 必填
    description?: string; // 可选
  };

  /** 创建标签响应 */
  type CreateTagResponse = BaseResponse & {
    data?: {
      id: string;
    };
  };

  /** 更新标签请求 */
  type UpdateTagRequest = {
    scope?: string;
    label?: string;
    description?: string;
  };

  /** 更新标签响应 */
  type UpdateTagResponse = BaseResponse;

  /** 删除标签响应 */
  type DeleteTagResponse = BaseResponse;

  /** 获取标签响应 */
  type GetTagResponse = BaseResponse & {
    data?: TagDefinition;
  };

  /** 标签列表查询请求 */
  type GetTagListRequest = PageParams & {
    scope: string; // 必填：tenant | user
    keyword?: string; // 模糊搜索
  };

  /** 标签列表数据 */
  type TagListData = {
    list: TagDefinition[];
  };

  /** 标签列表响应 */
  type GetTagListResponse = BaseResponse & {
    data?: TagListData;
    pageParams?: PageParams;
    // 兼容旧格式
    total?: number;
  };

  /** 获取标签选项请求 */
  type GetTagOptionsRequest = {
    scope: string;
    keyword?: string;
    limit?: number;
  };

  /** 获取标签选项响应 */
  type GetTagOptionsResponse = BaseResponse & {
    data?: {
      list: TagOption[];
    };
  };
}
