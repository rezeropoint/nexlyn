// 租户管理相关类型定义

declare namespace API {
  /** 租户完整信息 */
  type Tenant = {
    id: string; // 租户UUID（内部标识）
    tenantKey: string; // 业务标识符
    tenantName: string;
    description?: string;
    tags?: string[];
    contactEmail?: string;
    status: "active" | "inactive" | "deleted";
    createdAt: string;
    updatedAt: string;
    expiresAt?: string;
    createdBy: string;
    updatedBy: string;
    maxUsers: number;
  };

  /** 租户列表项（简要信息） */
  type TenantListItem = {
    id: string; // 租户UUID（内部标识）
    tenantKey: string; // 业务标识符
    tenantName: string;
  };

  /** 租户选项（用于下拉选择） */
  type TenantOption = {
    tenantId: string; // 租户ID（与id相同，用于权限管理）
    tenantKey: string; // 业务标识符
    tenantName: string;
    status: string;
  };

  /** 获取租户列表请求 */
  type GetTenantListRequest = {
    current?: number;
    pageSize?: number;
    tenantName?: string;
    status?: string;
  };

  /** 租户列表数据 */
  type TenantListData = {
    list: Tenant[];
  };

  /** 获取租户列表响应 */
  type GetTenantListResponse = BaseResponse & {
    data?: TenantListData;
    total?: number;
    current?: number;
    pageSize?: number;
  };

  /** 获取租户详情响应 */
  type GetTenantResponse = BaseResponse & {
    data?: Tenant;
  };

  /** 创建租户请求 */
  type CreateTenantRequest = {
    tenantKey: string;
    tenantName: string;
    description?: string;
    tags?: string[];
    contactEmail?: string;
    maxUsers?: number;
    expiresAt?: string;
    adminRoleKey: string; // 管理员角色标识（必填）
  };

  /** 创建租户响应 */
  type CreateTenantResponse = BaseResponse & {
    data?: {
      id: string;
      tenantKey: string;
    };
  };

  /** 更新租户请求 */
  type UpdateTenantRequest = {
    tenantName?: string;
    description?: string;
    tags?: string[];
    contactEmail?: string;
    maxUsers?: number;
    expiresAt?: string;
  };

  /** 更新租户响应 */
  type UpdateTenantResponse = BaseResponse;

  /** 更新租户状态请求 */
  type UpdateTenantStatusRequest = {
    status: "active" | "inactive" | "deleted";
  };

  /** 更新租户状态响应 */
  type UpdateTenantStatusResponse = BaseResponse;

  /** 删除租户响应 */
  type DeleteTenantResponse = BaseResponse;
}
