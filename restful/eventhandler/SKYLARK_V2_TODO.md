# Skylark V2 新增接口实施计划

> **版本**: go-skylark/v2@v2.3.1-beta20251119
> **创建日期**: 2025-11-13
> **最后更新**: 2025-11-19
> **状态**: 已完成 ✅（第一阶段10/10，第二阶段5/5，第三阶段3/3）

## 📋 目录

- [接口概览](#接口概览)
- [第一阶段：核心流程操作接口](#第一阶段核心流程操作接口)
- [第二阶段：用户和组织管理接口](#第二阶段用户和组织管理接口)
- [第三阶段：组织成员管理接口](#第三阶段组织成员管理接口)
- [核心数据模型](#核心数据模型)
- [实施建议](#实施建议)
- [注意事项](#注意事项)

---

## 接口概览

基于 go-skylark/v2 的 SkylarkEngine 接口，eventhandler 服务需要新增以下接口以提供完整的 Skylark 流程管理能力。

### V2 新增接口实施状态

| 接口分类 | 已完成 ✅ | 进行中 🚧 | 待实施 | 总计 |
|---------|----------|----------|--------|------|
| **第一阶段：流程操作** | **10** | **0** | **0** | **10** |
| **第二阶段：用户/组织同步** | **5** | **0** | **0** | **5** |
| **第三阶段：组织成员管理** | **3** | **0** | **0** | **3** |
| **总计** | **18** | **0** | **0** | **18** |

**已完成接口 ✅** (10个):
- UpdateFlowJourneyStatus（流程任务审批）
- AbortJourney（终止流程）
- GetFlowJourneyBySN（根据流程编号查询）
- GetFlowJourneyDetail（获取流程详情）
- GetJourneyMoments（获取审批历史时间线）
- GetCurrentProcessingUsers（获取当前处理人）
- GetUserAssignments（用户任务列表）
- GetProposedJourneys（用户发起的流程）
- SearchJourneys（高级搜索流程记录）
- GetFlowDetail（流程元数据）

**第二阶段已完成 ✅** (5个):
- eventsync gRPC 微服务（7个RPC：SyncUser、SyncOrganization、DeleteOrganization、BatchSync×2、GetSyncStatus×2、HealthCheck）
- backend CreateUserLogic（两阶段提交：先同步Skylark，失败则不创建本地）
- backend CreateOrganizationLogic（事务内同步：Skylark失败自动回滚）
- backend DeleteOrganizationLogic（异步删除Skylark组织）
- 配置管理（nexlyn-deploy/configs/backend/config.yaml：SkylarkSyncEnabled、AdminSyncClient）

**已排除**（由IoT引擎直接调用SDK，不暴露REST API）:
- CreateFlow（创建流程）
- CreateFormRow（创建表单行）

---

## 第一阶段：核心流程操作接口 ✅

> **状态**: 已完成（10/10）
> **完成日期**: 2025-11-17

### 已完成接口列表

| # | 接口名称 | REST API | 权限 |
|---|---------|----------|------|
| 1 | UpdateFlowJourneyStatus | `POST /api/v1/flows/:flowId/journeys/:journeyId/actions` | `flow_journey:approve` |
| 2 | AbortJourney | `DELETE /api/v1/flows/:flowId/journeys/:journeyId` | `flow:delete` |
| 3 | GetFlowJourneyBySN | `GET /api/v1/flows/:flowId/journeys/search-by-sn` | `flow:read` |
| 4 | GetFlowJourneyDetail | `GET /api/v1/flows/:flowId/journeys/:journeyId/detail` | `flow:read` |
| 5 | GetJourneyMoments | `GET /api/v1/flows/:flowId/journeys/:journeyId/moments` | `flow:read` |
| 6 | GetCurrentProcessingUsers | `GET /api/v1/flows/:flowId/journeys/:journeyId/processing-users` | `flow:read` |
| 7 | SearchJourneys | `POST /api/v1/flows/:flowId/journeys/search` | `flow:read` |
| 8 | GetUserAssignments | `GET /api/v1/my-assignments` | 无（仅自己） |
| 9 | GetProposedJourneys | `GET /api/v1/my-proposed-journeys` | 无（仅自己） |
| 10 | GetFlowDetail | `GET /api/v1/flows/:flowId/detail` | `flow:read` |

### 实施总结

- **实现模式**: JWT认证 → CasbinX权限验证 → SDK调用 → 类型转换 → 返回响应
- **用户ID类型**: 统一使用 `string` 类型（UUID）
- **SDK版本**: v2.1.3（可选参数简化为 `string` 类型）
- **错误处理**: 统一通过 `svc.HandleSkylarkError()` 处理

### 已排除接口（由IoT引擎直接调用SDK）
- CreateFlow（创建流程）
- CreateFormRow（创建表单行）

---

## 第二阶段：用户和组织管理接口 ✅

> **优先级**: ⭐⭐ 中
> **实际工作量**: 3天
> **状态**: 已完成（2025-11-17）

### ✅ 已完成：eventsync gRPC 微服务 (2025-11-17)

#### 服务架构
- **位置**: `service/eventsync/`
- **监听端口**: 9998
- **Docker 服务名**: `eventsync`
- **技术栈**: go-zero gRPC + go-skylark/v2 AdminEngine

#### 已实现的 RPC 方法（6个）
1. **SyncUser** - 同步单个用户到 Skylark
2. **SyncOrganization** - 同步组织（自动判断根/子组织）
3. **DeleteOrganization** - 删除组织同步
4. **BatchSyncUsers** - 批量同步用户（接收完整用户数据）
5. **BatchSyncOrganizations** - 批量同步组织（接收完整组织数据，按层级排序）
6. **HealthCheck** - 健康检查

#### 关键设计点
- ✅ 不查询数据库，只调用 AdminEngine
- ✅ 批量同步由调用方传递完整数据
- ✅ ID 类型转换（int → string）
- ✅ 错误处理（同步失败不抛异常，返回详细错误信息）
- ✅ 支持配置开关（SkylarkSyncEnabled）

#### backend 集成状态（全部完成 ✅）
- ✅ Config 已添加 `SkylarkSyncEnabled` 和 `AdminSyncClient`
- ✅ ServiceContext 已初始化 gRPC 客户端
- ✅ 配置文件已更新（`eventsync:9998`，10秒超时）
- ✅ CreateUserLogic 集成（两阶段提交：先同步Skylark，失败则不创建本地）
- ✅ CreateOrganizationLogic 集成（事务内同步：Skylark失败自动回滚）
- ✅ DeleteOrganizationLogic 集成（异步删除Skylark组织）
- ✅ GetUserSyncStatus/GetOrgSyncStatus（使用v2.2.0新增接口）

### ✅ 已完成：backend 业务逻辑集成

#### 1. CreateUser 集成（异步同步）
**位置**: `restful/backend/internal/logic/user/createuserlogic.go`

**实施步骤**:
```go
// 1. 创建本地用户（现有逻辑）
userID, err := l.createLocalUser(req)
if err != nil {
    return nil, err
}

// 2. 异步同步到 Skylark（新增）
if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
    go l.syncUserToSkylark(currentUser.TenantID, userID, req)
}

return &types.CreateUserResponse{...}, nil
```

**辅助方法**:
```go
func (l *CreateUserLogic) syncUserToSkylark(tenantID, userID string, req *types.CreateUserRequest) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := l.svcCtx.AdminSyncClient.SyncUser(ctx, &eventsync.SyncUserReq{
        TenantId:    tenantID,
        LocalUserId: userID,
        Name:        req.UserName,
        Identifier:  req.Email,     // 可选
        Phone:       req.Phone,     // 可选
    })
    if err != nil {
        logx.Errorf("同步用户到Skylark失败: %v (本地用户已创建: %s)", err, userID)
    }
}
```

#### 2. CreateOrganization 集成（异步同步）
**位置**: `restful/backend/internal/logic/organization/createorganizationlogic.go`

**实施步骤**:
```go
// 1. 创建本地组织（现有逻辑）
orgID, err := l.createLocalOrganization(req)
if err != nil {
    return nil, err
}

// 2. 异步同步到 Skylark
if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
    go l.syncOrganizationToSkylark(currentUser.TenantID, orgID, req)
}
```

**辅助方法**:
```go
func (l *CreateOrganizationLogic) syncOrganizationToSkylark(tenantID, orgID string, req *types.CreateOrganizationRequest) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err := l.svcCtx.AdminSyncClient.SyncOrganization(ctx, &eventsync.SyncOrganizationReq{
        TenantId:          tenantID,
        LocalOrgId:        orgID,
        ParentLocalOrgId:  req.ParentOrgId,  // 空字符串表示根组织
        Name:              req.OrgName,
        Description:       req.Description,
        FounderId:         req.FounderId,
    })
    if err != nil {
        logx.Errorf("同步组织到Skylark失败: %v (本地组织已创建: %s)", err, orgID)
    }
}
```

#### 3. DeleteOrganization 集成
**位置**: `restful/backend/internal/logic/organization/deleteorganizationlogic.go`

**实施步骤**:
```go
// 1. 删除本地组织（现有逻辑）
err := l.deleteLocalOrganization(orgID)
if err != nil {
    return err
}

// 2. 同步删除 Skylark 组织（可选择同步或异步）
if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
    l.deleteSkylarkOrganization(currentUser.TenantID, orgID)
}
```

#### 4. 批量同步接口（历史数据迁移）
**建议**: 在 backend 添加管理员接口，调用 eventsync 的批量同步方法

**位置**: `restful/backend/internal/logic/admin/syncskylarklogic.go`（新建）

**实施步骤**:
1. 查询所有用户/组织
2. 构造 `UserInfo` 和 `OrganizationInfo` 列表
3. 调用 `BatchSyncUsers` 和 `BatchSyncOrganizations`
4. 返回同步结果统计

---

## 第三阶段：组织成员管理接口 ✅

> **状态**: 已完成（3/3）
> **完成日期**: 2025-11-19
> **SDK版本**: go-skylark/v2@v2.3.1-beta20251119

### 已完成接口列表

基于 go-skylark/v2 v2.3.1 新增的 AdminEngine 接口，实现了完整的组织成员管理功能：

| # | SDK接口 | eventsync RPC | backend集成 | 同步策略 |
|---|---------|--------------|------------|---------|
| 1 | UpdateOrganization | UpdateOrganization | UpdateOrganizationLogic | 事务内同步 |
| 2 | AddMember | AddMember | AddOrganizationMemberLogic | 事务内同步 |
| 3 | RemoveMember | RemoveMember | RemoveOrganizationMemberLogic | 事务内同步 |

### 实施详情

#### 1. UpdateOrganization - 更新组织信息
**SDK接口**: `AdminEngine.UpdateOrganization(ctx, *core.UpdateOrganizationRequest)`

**功能**:
- 支持部分更新（名称、描述、管理员）
- 使用分布式锁确保原子性
- SDK内部封装多次API调用（查询旧管理员 → 添加新管理员 → 删除旧管理员）

**eventsync RPC**:
```protobuf
message UpdateOrganizationReq {
  string tenant_id = 1;
  string local_org_id = 2;
  string name = 3;
  string description = 4;
  string manager_id = 5;
  bool update_name = 6;
  bool update_description = 7;
  bool update_manager = 8;
}
```

**backend集成**: `restful/backend/internal/logic/organization/updateorganizationlogic.go`
- **集成方式**: 事务内同步（Skylark失败则回滚本地数据库）
- **变更检测**: 对比旧值，只同步变化的字段
- **超时时间**: 10秒
- **错误处理**: Skylark失败回滚整个事务，确保数据一致性

#### 2. AddMember - 添加成员到组织
**SDK接口**: `AdminEngine.AddMember(ctx, tenantID, localOrgID, localMemberID)`

**功能**:
- SDK内部自动转换本地用户ID为远程用户ID
- 成功后自动清理相关缓存

**eventsync RPC**:
```protobuf
message AddMemberReq {
  string tenant_id = 1;
  string local_org_id = 2;
  string local_member_id = 3;
}
```

**backend集成**: `restful/backend/internal/logic/organization/addorganizationmemberlogic.go`
- **集成方式**: 事务内同步（Skylark失败则回滚本地数据库）
- **超时时间**: 10秒
- **错误处理**: Skylark失败回滚整个事务，确保数据一致性

#### 3. RemoveMember - 从组织移除成员
**SDK接口**: `AdminEngine.RemoveMember(ctx, tenantID, localOrgID, localMemberID)`

**功能**:
- SDK内部自动转换本地用户ID为远程用户ID
- 成功后自动清理相关缓存

**eventsync RPC**:
```protobuf
message RemoveMemberReq {
  string tenant_id = 1;
  string local_org_id = 2;
  string local_member_id = 3;
}
```

**backend集成**: `restful/backend/internal/logic/organization/removeorganizationmemberlogic.go`
- **集成方式**: 事务内同步（Skylark失败则回滚本地数据库软删除）
- **超时时间**: 10秒
- **错误处理**: Skylark失败回滚整个事务，确保数据一致性

### 同步策略对比

| 操作 | 本地操作 | Skylark同步 | 失败影响 |
|-----|---------|------------|---------|
| UpdateOrganization | UPDATE + 事务 | 事务内同步 | 回滚本地更新 |
| AddMember | INSERT + 事务 | 事务内同步 | 回滚本地插入 |
| RemoveMember | UPDATE（软删除）+ 事务 | 事务内同步 | 回滚本地软删除 |

**统一策略**: 所有组织成员管理操作均采用**事务内同步**，确保本地数据库和 Skylark 平台的数据一致性。

### 实施文件清单

**eventsync服务** (7个文件):
1. ✅ `service/eventsync/eventsync.proto` - 新增3个RPC定义和6个Message
2. ✅ `service/eventsync/pb/eventsync.pb.go` - 自动生成
3. ✅ `service/eventsync/pb/eventsync_grpc.pb.go` - 自动生成
4. ✅ `service/eventsync/internal/logic/updateorganizationlogic.go` - 新建
5. ✅ `service/eventsync/internal/logic/addmemberlogic.go` - 新建
6. ✅ `service/eventsync/internal/logic/removememberlogic.go` - 新建
7. 🔄 `service/eventsync/internal/server/adminsyncserver.go` - 可能需要手动注册

**backend服务** (3个文件):
8. ✅ `restful/backend/internal/logic/organization/updateorganizationlogic.go` - 新增异步同步
9. ✅ `restful/backend/internal/logic/organization/addorganizationmemberlogic.go` - 新增事务内同步
10. ✅ `restful/backend/internal/logic/organization/removeorganizationmemberlogic.go` - 新增事务内同步

### 技术要点

#### 错误处理模式
```go
// eventsync RPC - 统一返回格式（不抛异常）
if err != nil {
    return &pb.AddMemberResp{
        Success: false,
        Message: err.Error(),
        ErrorCode: "ADD_MEMBER_FAILED",
    }, nil
}
```

```go
// backend 事务内同步 - 失败则回滚
if syncErr != nil || !syncResp.Success {
    logx.Error("同步失败，回滚事务")
    return fmt.Errorf("添加成员到 Skylark 失败: %v", syncErr)
}
```

#### 统一的事务内同步策略
所有组织成员管理操作（UpdateOrganization、AddMember、RemoveMember）均采用**事务内同步**：

**优势**：
- ✅ **数据一致性**: Skylark 失败则回滚本地数据库，避免数据不一致
- ✅ **简化逻辑**: 无需异步补偿机制
- ✅ **错误处理**: 用户立即得到同步失败的反馈

**事务流程**：
```go
err = l.svcCtx.DBConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
    // 1. 更新本地数据库
    _, err := session.ExecCtx(ctx, updateQuery, ...)
    if err != nil {
        return err
    }

    // 2. 同步到 Skylark（超时 10 秒）
    if l.svcCtx.Config.SkylarkSyncEnabled && l.svcCtx.AdminSyncClient != nil {
        syncResp, syncErr := l.svcCtx.AdminSyncClient.XXX(ctx, req)
        if syncErr != nil || !syncResp.Success {
            return fmt.Errorf("同步到 Skylark 失败: %v", syncErr) // 回滚
        }
    }

    return nil // 提交事务
})
```

#### 配置开关
所有Skylark同步均受 `SkylarkSyncEnabled` 配置控制：
```yaml
# restful/backend/etc/config.yaml
SkylarkSyncEnabled: false  # 默认关闭
```

---

## 核心数据模型

### TypedValue 字段类型
- `string`, `number`, `boolean`, `date`（ISO 8601）, `array`, `object`, `file`

### 流程状态常量
- **状态**: `processing`（进行中）, `completed`（已完成）, `aborted`（已终止）, `stashed`（已搁置）
- **任务类别**: `proposed`（我发起的）, `processed`（待我处理）, `cc`（抄送我的）
- **操作类型**: `approve`（通过）, `refuse`（回退）, `transfer`（转交）, `cancel`（撤销）

---

## 实施建议

### 分阶段实施路线图

#### ✅ 阶段一：核心流程操作（已完成 10/10）
- ✅ 10个接口全部实现
- 🚧 单元测试和前端联调（待进行）

#### 🚧 阶段二：用户和组织同步（基础设施已完成）
- ✅ eventsync gRPC 微服务（6个RPC方法）
- ✅ backend 配置集成
- 🚧 3个业务逻辑集成（CreateUser、CreateOrganization、DeleteOrganization）

#### 阶段三：高级功能
- ✅ GetFlowDetail 已完成
- ❌ 其他接口不需要实现或已合并

### 技术依赖和前置条件

#### 已满足 ✅
- [x] go-skylark/v2@v2.1.3 已升级
- [x] 平台配置表支持 APIToken 字段
- [x] SkylarkEngine 已集成

#### 待准备
- [ ] 平台配置中填写 APIBaseURL 和 APIToken
- [ ] 测试环境部署 Skylark 实例
- [ ] 前端页面开发（流程审批UI、待办列表等）

### 认证策略
**推荐方案**：统一使用管理员 Token，SDK 内部通过用户ID映射实现权限隔离

---

## 注意事项

### ⚠️ 用户映射管理
- SDK 自动处理用户ID映射（localUserID → remoteUserID）
- 映射不存在时返回 `ErrUserMappingNotFound` 错误
- 用户映射由第二阶段的同步接口创建

### ⚠️ 权限验证
- 三步验证：JWT解析 → CasbinX功能权限 → SDK数据级权限
- SDK内部验证用户是否为任务处理人

### ⚠️ 错误处理
- 统一通过 `svc.HandleSkylarkError()` 转换SDK错误
- 错误映射定义在 `helpers.go` 中

### ⚠️ 性能优化
- FlowDetail、用户映射、组织映射使用Redis缓存
- SDK自动批量查询，避免N+1问题

---

## 参考信息

### 相关文档
- [go-skylark/v2 README](https://github.com/rezeropoint/go-skylark/blob/v2/README.md)
- [Skylark API 文档](../pkg/skylarkq/Skylark流程API文档.md)
- [Skylark 用户 API](../pkg/skylarkq/Skylark用户API文档.md)
- [Skylark 组织 API](../pkg/skylarkq/Skylark组织API文档.md)

### 版本信息
- **go-skylark/v2**: v2.1.3-beta20251117
- **go-zero**: v1.9.2
- **eventsync服务端口**: 9998


## 实施检查清单

### ✅ 阶段一完成标准（已完成）
- [x] 10个接口的API定义、Handler、Logic全部实现
- [x] 类型转换函数和权限资源定义完成
- [x] 用户ID类型修正（int64 → string UUID）
- [x] SDK v2.1.3适配完成
- [ ] 单元测试（待补充）
- [ ] 前端联调（待进行）

### 阶段二完成标准 ✅
- [x] eventsync gRPC 微服务已创建（7个RPC方法）
- [x] backend Config 和 ServiceContext 已集成
- [x] 配置文件已更新
- [x] 编译测试通过
- [x] CreateUserLogic 两阶段提交集成
- [x] CreateOrganizationLogic 事务内同步集成
- [x] DeleteOrganizationLogic 异步删除集成
- [x] GetUserSyncStatus/GetOrgSyncStatus 真实实现（v2.2.0）

---

**最后更新**: 2025-11-17
**维护者**: 开发团队
**状态**: ✅ 全部完成（第一阶段10/10，第二阶段5/5）
**版本**: go-skylark/v2@v2.2.0-beta20251117

