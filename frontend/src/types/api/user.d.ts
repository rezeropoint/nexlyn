// 用户管理相关类型定义

declare namespace API {
  /** 租户信息 */
  type TenantInfo = {
    tenantKey: string;
    tenantId: string;
    tenantName: string;
  };

  /** 用户完整信息 */
  type User = {
    // 基础身份信息
    id: string; // 用户ID（用于URL路径参数）
    userKey: string;
    userName: string; // 前端继续使用userName，但对应后端的userName
    email: string;
    phone: string;
    // 个人信息
    name: string;
    avatar: string;
    signature: string;
    title: string;
    // 多租户信息
    tenantInfo: TenantInfo;
    // 组织架构信息
    organizationIds?: string[]; // 用户关联的组织ID列表
    // 用户角色
    role: RoleKey; // super_admin, admin, user
    // 权限信息（基于Casbin的完整权限数据）
    permissions: PermissionResource[]; // 用户具体权限列表
    roles: RoleKey[]; // 用户所有角色列表
    isActive: boolean; // 用户是否活跃
    // 标签和地理信息
    tags: UserTag[];
    geographic: UserGeographic;
    country: string;
    address: string;
    // 状态和时间戳
    status: string; // 'active'-活跃，'inactive'-禁用，'deleted'-删除
    createdAt: string;
    updatedAt: string;
    lastLoginAt: string;
  };

  /** 用户简要信息（用于列表显示） */
  type UserBrief = {
    id: string; // 用户ID（用于URL路径参数）
    userKey: string;
    userName: string; // 前端继续使用userName，但对应后端的userName
    name: string;
    email: string;
    avatar: string;
    role: RoleKey;
    status: string;
    tenantInfo: TenantInfo;
    isActive: boolean;
    // 添加缺失的字段
    tags?: UserTag[];
    createdAt?: number;
    updatedAt?: number;
    lastLoginAt?: number;
  };

  /** 兼容性：保留 CurrentUser 类型 */
  type CurrentUser = User;

  // ===== 用户管理API类型 =====

  /** 获取当前用户响应 */
  type GetCurrentUserResponse = BaseResponse & {
    data?: User;
  };

  /** 用户列表查询请求 */
  type GetUserListRequest = PageParams & {
    // 按用户ID列表查询（优先级最高，传入时忽略其他搜索条件）
    ids?: string[];
    // 搜索条件
    name?: string;
    email?: string;
    userName?: string;
    status?: string;
    // 多租户过滤
    tenantId?: string;
    // 标签过滤（使用标签ID，包含任一）
    tagIds?: string[];
  };

  /** 用户列表数据 */
  type UserListData = {
    list?: UserBrief[];
  };

  /** 用户列表响应 */
  type GetUserListResponse = BaseResponse &
    PageParams & {
      data?: UserListData;
    };

  /** 根据ID获取用户响应 */
  type GetUserByIdResponse = BaseResponse & {
    data?: User;
  };

  /** 创建用户请求 */
  type CreateUserRequest = {
    // 基础信息（必需）
    name: string;
    email: string;
    userName: string;
    // 租户信息（必需，指定用户所属租户）
    tenantId: string;
    // 个人信息（可选）
    phone?: string;
    title?: string;
    avatar?: string;
    signature?: string;
    country?: string;
    address?: string;
    // 组织信息（可选）
    role?: RoleKey;
    // 扩展信息（使用标签ID）
    tagIds?: string[];
    geographic?: UserGeographic;
  };

  /** 创建用户响应数据 */
  type CreateUserData = {
    userKey: string;
    userName: string; // 对应后端的userName
    password: string;
  };

  /** 创建用户响应 */
  type CreateUserResponse = BaseResponse & {
    data?: CreateUserData;
  };

  /** 更新用户请求 */
  type UpdateUserRequest = {
    // 基础信息
    name?: string;
    email?: string;
    userName?: string;
    phone?: string;
    // 个人信息
    title?: string;
    avatar?: string;
    signature?: string;
    country?: string;
    address?: string;
    // 扩展信息（使用标签ID，覆盖式更新）
    tagIds?: string[];
    geographic?: UserGeographic;
  };

  /** 更新用户响应 */
  type UpdateUserResponse = BaseResponse;

  /** 更新用户状态请求 */
  type UpdateUserStatusRequest = {
    status?: string; // 用户状态：'active'-活跃，'inactive'-禁用，'deleted'-删除
    role?: string; // 用户角色
  };

  /** 删除用户响应 */
  type DeleteUserResponse = BaseResponse;

  /** 重置密码请求 */
  type ResetPasswordRequest = {
    password: string;
  };

  /** 重置密码响应 */
  type ResetPasswordResponse = BaseResponse;

  /** 修改当前用户密码请求 */
  type ChangePasswordRequest = {
    oldPassword: string; // 旧密码
    newPassword: string; // 新密码
  };

  /** 修改当前用户密码响应 */
  type ChangePasswordResponse = BaseResponse;

  /** 更新当前用户信息请求（用户修改自己的资料，无需user:write权限） */
  type UpdateCurrentUserRequest = {
    // 基础信息
    name?: string; // 用户姓名
    email?: string; // 电子邮件
    phone?: string; // 电话号码
    // 个人信息
    title?: string; // 职位
    avatar?: string; // 用户头像 URL
    signature?: string; // 个人签名
  };

  /** 更新当前用户信息响应 */
  type UpdateCurrentUserResponse = BaseResponse;

  /** 更新用户状态响应 */
  type UpdateUserStatusResponse = BaseResponse;

  // ===== 认证相关类型 =====

  /** 登录请求 */
  type LoginParams = {
    userName?: string; // 与后端保持一致
    password?: string;
    autoLogin?: boolean;
    type?: string;
  };

  /** 登录响应 */
  type LoginResponse = BaseResponse & {
    token?: string;
  };

  // ===== 头像上传相关类型 =====

  /** 上传头像响应数据 */
  type UploadAvatarData = {
    avatarUrl: string; // 头像URL
  };

  /** 上传头像响应 */
  type UploadAvatarResponse = BaseResponse & {
    data?: UploadAvatarData;
  };
}
