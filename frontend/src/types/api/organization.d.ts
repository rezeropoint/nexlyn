// 组织管理相关API类型定义
declare namespace API {
  // ===== 组织实体类型 =====

  /** 组织实体 */
  interface Organization {
    id: string; // 组织UUID（内部标识）
    code: string; // 组织代码（业务标识）
    name: string; // 组织名称
    type: string; // 组织类型（company/department/team等）
    description: string; // 描述
    path: string; // 路径编码（/root/parent_id/current_id/）
    parentId: string; // 父组织ID
    level: number; // 层级（1-10）
    sortOrder: number; // 排序
    managerId: string; // 负责人ID
    managerName: string; // 负责人姓名
    memberCount: number; // 成员数量
    status: string; // 状态（active/inactive/deleted）
    tenantId: string; // 租户ID
    createdBy: string; // 创建者
    updatedBy: string; // 更新者
    createdAt: string; // 创建时间
    updatedAt: string; // 更新时间
  }

  /** 组织树结构 */
  interface OrganizationTree extends Organization {
    children: OrganizationTree[]; // 子组织
  }

  /** 简化组织信息（用于选项和列表） */
  interface OrganizationBrief {
    id: string; // 组织UUID
    code: string; // 组织代码
    name: string; // 组织名称
    type: string; // 组织类型
    path: string; // 路径编码
    parentId: string; // 父组织ID
    level: number; // 层级
    managerName: string; // 负责人姓名
    memberCount: number; // 成员数量
    status: string; // 状态
  }

  /** 组织选项（用于下拉框） */
  interface OrganizationOption {
    id: string; // 组织UUID
    code: string; // 组织代码
    name: string; // 组织名称
    path: string; // 路径编码
    level: number; // 层级
    status: string; // 状态
  }

  /** 用户组织关系 */
  interface UserOrgRelation {
    id: string; // 关系ID
    userId: string; // 用户ID
    userKey: string; // 用户Key
    userName: string; // 用户名
    name: string; // 用户姓名
    email: string; // 用户邮箱
    orgId: string; // 组织ID
    orgCode: string; // 组织代码
    orgName: string; // 组织名称
    orgType: string; // 组织类型
    relationType: string; // 关系类型（manager/member）
    positionTitle: string; // 职位标题
    isPrimary: boolean; // 是否主要关系
    status: string; // 关系状态
    joinedAt: string; // 加入时间
  }

  // ===== 请求类型 =====

  /** 获取组织树请求 */
  interface GetOrganizationTreeRequest {
    tenantId?: string; // 租户ID过滤（超级管理员可指定）
    rootId?: string; // 根节点ID（默认为租户根组织）
  }

  /** 获取组织详情请求 */
  interface GetOrganizationRequest {
    id: string; // 组织ID
  }

  /** 组织列表查询请求 */
  interface GetOrganizationListRequest extends PageParamsRequest {
    code?: string; // 组织代码
    name?: string; // 组织名称模糊查询
    type?: string; // 组织类型
    status?: string; // 状态过滤
    parentId?: string; // 父组织ID
    level?: number; // 层级过滤
    tenantId?: string; // 租户ID（超级管理员可指定）
  }

  /** 获取组织选项请求 */
  interface GetOrganizationOptionsRequest {
    tenantId?: string; // 租户ID过滤（超级管理员可指定）
    keyword?: string; // 关键字搜索（组织名称、代码）
    parentId?: string; // 父组织ID过滤
    type?: string; // 组织类型过滤
    limit?: number; // 限制数量（默认50）
  }

  /** 创建组织请求 */
  interface CreateOrganizationRequest {
    code: string; // 组织代码（必填）
    name: string; // 组织名称（必填）
    type?: string; // 组织类型（默认department）
    description?: string; // 描述
    parentId?: string; // 父组织ID（为空则创建为根组织）
    sortOrder?: number; // 排序（默认0）
    managerId?: string; // 负责人ID
    status?: string; // 状态（默认active）
  }

  /** 更新组织请求 */
  interface UpdateOrganizationRequest {
    id: string; // 组织ID
    name?: string; // 组织名称
    type?: string; // 组织类型
    description?: string; // 描述
    sortOrder?: number; // 排序
    managerId?: string; // 负责人ID
    status?: string; // 状态
  }

  /** 删除组织请求 */
  interface DeleteOrganizationRequest {
    id: string; // 组织ID
    forceDelete?: boolean; // 强制删除（级联删除子组织，默认false）
  }

  /** 移动组织请求 */
  interface MoveOrganizationRequest {
    id: string; // 组织ID
    newParentId: string; // 新父组织ID（null表示移动到根级）
    newSortOrder?: number; // 新排序（可选）
  }

  /** 获取组织成员请求 */
  interface GetOrganizationMembersRequest extends PageParamsRequest {
    id: string; // 组织ID
    keyword?: string; // 关键字搜索（用户名、姓名、邮箱）
    relationType?: string; // 关系类型过滤（manager/member）
    isPrimary?: boolean; // 是否主要关系过滤
  }

  /** 添加组织成员请求 */
  interface AddOrganizationMemberRequest {
    id: string; // 组织ID
    userId: string; // 用户ID（必填）
    relationType?: string; // 关系类型（默认member）
    positionTitle?: string; // 职位标题
    isPrimary?: boolean; // 是否主要关系（默认false）
  }

  /** 移除组织成员请求 */
  interface RemoveOrganizationMemberRequest {
    orgId: string; // 组织ID
    userId: string; // 用户ID
  }

  /** 更新成员关系请求 */
  interface UpdateOrganizationMemberRequest {
    orgId: string; // 组织ID
    userId: string; // 用户ID
    relationType?: string; // 关系类型
    positionTitle?: string; // 职位标题
    isPrimary?: boolean; // 是否主要关系
    status?: string; // 关系状态
  }

  /** 获取用户所属组织请求 */
  interface GetUserOrganizationsRequest {
    id: string; // 用户ID
    isPrimary?: boolean; // 是否只返回主要关系
  }

  // ===== 响应类型 =====

  /** 获取组织树响应 */
  interface GetOrganizationTreeResponse extends BaseResponse {
    data?: OrganizationTree[];
  }

  /** 获取组织详情响应 */
  interface GetOrganizationResponse extends BaseResponse {
    data?: Organization;
  }

  /** 组织列表数据 */
  interface OrganizationListData {
    list: OrganizationBrief[];
  }

  /** 组织列表响应 */
  interface GetOrganizationListResponse extends BaseResponse, PageParams {
    data?: OrganizationListData;
  }

  /** 获取组织选项响应 */
  interface GetOrganizationOptionsResponse extends BaseResponse {
    data?: {
      list: OrganizationOption[];
    };
  }

  /** 创建组织响应 */
  interface CreateOrganizationResponse extends BaseResponse {
    data?: {
      id: string; // 新创建的组织ID
      code: string; // 组织代码
      path: string; // 路径编码
    };
  }

  /** 更新组织响应 */
  interface UpdateOrganizationResponse extends BaseResponse {}

  /** 删除组织响应 */
  interface DeleteOrganizationResponse extends BaseResponse {
    data?: {
      deletedCount: number; // 删除的组织数量（包含子组织）
    };
  }

  /** 移动组织响应 */
  interface MoveOrganizationResponse extends BaseResponse {
    data?: {
      newPath: string; // 新路径编码
    };
  }

  /** 组织成员列表数据 */
  interface OrganizationMembersData {
    list: UserOrgRelation[];
  }

  /** 获取组织成员响应 */
  interface GetOrganizationMembersResponse extends BaseResponse, PageParams {
    data?: OrganizationMembersData;
  }

  /** 添加组织成员响应 */
  interface AddOrganizationMemberResponse extends BaseResponse {}

  /** 移除组织成员响应 */
  interface RemoveOrganizationMemberResponse extends BaseResponse {}

  /** 更新成员关系响应 */
  interface UpdateOrganizationMemberResponse extends BaseResponse {}

  /** 获取用户所属组织响应 */
  interface GetUserOrganizationsResponse extends BaseResponse {
    data?: {
      list: UserOrgRelation[];
    };
  }
}
