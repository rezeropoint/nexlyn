# 门禁打卡系统集成方案（Access 项目迁移）

## 📋 文档概述

本文档描述将独立的 Access 门禁打卡项目迁移到 LynxGraph 逻辑引擎的方案。

**当前状态**: ⏳ Phase 4 进行中

**最后更新**: 2025-01-25

---

## 🎯 项目背景

将独立的 Access 门禁打卡项目（`D:\Documents\Code\access`）迁移到 LynxGraph 逻辑引擎，实现：
- **可视化配置**：通过逻辑图编辑器配置打卡流程，无需修改代码
- **灵活扩展**：轻松添加新的打卡规则、假期策略、通知渠道
- **统一管理**：与 IoT 设备数据共用同一套逻辑引擎

---

## 📊 原 Access 系统功能

| 功能 | 说明 |
|------|------|
| 人脸识别接收 | 接收摄像头人脸识别数据 |
| 智能开门 | 识别成功后调用 Webhook 开门 |
| 工作日判断 | 检查假期/周末/补班日 |
| 打卡去重 | Redis 缓存当天打卡记录 |
| 迟到判断 | 上班时间后识别 → 检查请假/外勤 → 迟到打卡 |
| 流程触发 | 调用 Skylark 创建打卡/迟到记录 |
| 定时通知 | 每天 9:31 通知未打卡员工 |

### 原系统数据流

```
外部系统(人脸识别摄像头)
    ↓ (HTTP POST /api/faceid)
    ↓ FaceIdRequest
    │  ├─ Result.Description  (识别编号)
    │  ├─ Result.Tags[0]      (识别到的人名)
    │  ├─ Result.Type         (识别类型)
    │  └─ Summary, TimeStamp   (摘要和时间戳)
    ↓
Handler → Logic 层核心处理
    ├─ Redis 查询打卡缓存
    ├─ PostgreSQL 查询（员工/请假/外勤）
    ├─ 假期服务 API 查询
    ├─ Webhook 请求（开门/音效）
    └─ Skylark 引擎（创建流程）
```

### 时间阶段判断逻辑

| 时间段 | 条件 | 处理方式 | 触发流程 |
|--------|------|--------|--------|
| 工作日上班前 | now < 09:31 | 正常打卡 + 缓存 | Flow 135 |
| 工作日上班后(无请假/外勤) | 09:31 ≤ now ≤ 17:31 | 迟到打卡 + 缓存 | Flow 189 |
| 工作日上班后(有请假/外勤) | 09:31 ≤ now ≤ 17:31 | 无需打卡 | 无 |
| 下班后 | now > 17:31 | 无需打卡 | 无 |
| 非工作日 | 周末或假期 | 无需打卡 | 无 |
| 当天已打过卡 | Redis 缓存存在 | 无需打卡 | 无 |

---

## 🏗️ 逻辑图架构设计

```
人脸识别信息原子（触发）
    ↓
① HttpRequest（开门 Webhook）──→ 立即开门
    ↓
② HolidayCheck（假期检查）
    ├─ 非工作日 ──→ 终止（无需打卡）
    └─ 工作日 ──→ 继续
    ↓
③ CacheCheck（打卡缓存检查）
    ├─ 已打卡 ──→ 终止（重复识别）
    └─ 未打卡 ──→ 继续
    ↓
④ TimeWindowCheck（时间窗口判断）
    ├─ 上班前（< 9:31）──→ 正常打卡分支
    └─ 上班后（≥ 9:31）──→ 迟到检查分支
    ↓
┌───────────────────┬────────────────────────┐
│   正常打卡分支      │      迟到检查分支         │
│         ↓         │           ↓             │
│  CacheSet（缓存）   │  QueryDatabase（请假）   │
│         ↓         │           ↓             │
│  SkylarkFlow       │  QueryDatabase（外勤）   │
│  (Flow 135)       │           ↓             │
│         ↓         │    Switch（条件路由）      │
│  HttpRequest      │    ├─ 有请假/外勤 → 终止   │
│  (播放音效)        │    └─ 无 → CacheSet      │
│                   │              ↓           │
│                   │    SkylarkFlow(Flow 189) │
└───────────────────┴────────────────────────┘
```

---

## 📝 信息原子类型定义

```yaml
name: "face_recognition"
description: "人脸识别事件"
fields:
  - name: "person_name"
    type: "string"
    jsonPath: "$.Result.Tags[0]"
    description: "识别到的人员姓名"
  - name: "recognition_id"
    type: "string"
    jsonPath: "$.Result.Description"
    description: "识别编号"
  - name: "timestamp"
    type: "integer"
    jsonPath: "$.TimeStamp"
    description: "识别时间戳（毫秒）"
```

---

## 🧱 需要实现的积木

| 积木名称 | 类型 | 优先级 | 说明 |
|---------|------|--------|------|
| **Switch** | 条件路由 | 🔴 P0 | 多分支条件判断，支持表达式 |
| **HttpRequest** | 外部调用 | 🔴 P0 | HTTP GET/POST 请求（Webhook） |
| **CacheCheck** | 缓存查询 | 🔴 P0 | Redis 缓存存在性检查 |
| **CacheSet** | 缓存写入 | 🔴 P0 | Redis 缓存写入（带 TTL） |
| **TimeWindowCheck** | 时间判断 | 🔴 P0 | 检查当前时间是否在指定窗口内 |
| **HolidayCheck** | 假期判断 | 🟡 P1 | 调用假期 API 判断工作日 |
| **QueryDatabase** | 数据库查询 | 🟡 P1 | PostgreSQL 查询（请假/外勤） |
| **SkylarkFlow** | 流程触发 | 🟡 P1 | 调用 Skylark 创建流程 |

---

## 📐 积木详细设计

### 1. Switch（条件路由）⭐ 核心积木

**功能**: 根据条件表达式路由到不同分支

**配置参数**:
```yaml
conditions:
  - name: "on_leave"
    expression: "context.is_on_leave == true"
    targetNode: "terminate_leave"
  - name: "on_field_work"
    expression: "context.is_on_field_work == true"
    targetNode: "terminate_field"
  - name: "default"
    expression: "true"
    targetNode: "late_clock_in"
```

**表达式支持**:
- 变量引用: `context.{key}`, `atom.{field}`, `env.{name}`
- 比较运算: `==`, `!=`, `>`, `<`, `>=`, `<=`
- 逻辑运算: `&&`, `||`, `!`
- 函数调用: `contains()`, `startsWith()`, `isEmpty()`

---

### 2. HttpRequest（HTTP 请求）

**功能**: 发送 HTTP 请求到外部系统

**配置参数**:
```yaml
method: "POST"  # GET/POST/PUT/DELETE
url: "http://192.168.6.95:8123/api/webhook/-O6Wb2..."
headers:
  Content-Type: "application/json"
body: '{"action": "open_door", "person": "{{atom.person_name}}"}'
timeout: 5000  # 毫秒
saveResponseTo: "door_response"  # 可选，保存响应到 GraphContext
ignoreError: true  # 可选，忽略错误继续执行
```

---

### 3. CacheCheck / CacheSet（缓存操作）

**CacheCheck 配置**:
```yaml
keyPattern: "face_record_{{atom.person_name}}_{{date}}"
saveExistsTo: "already_clocked_in"  # 保存结果到 GraphContext
terminateIfExists: true  # 可选，存在则终止执行
```

**CacheSet 配置**:
```yaml
keyPattern: "face_record_{{atom.person_name}}_{{date}}"
value: "{{timestamp}}"
ttl: 64800  # 18 小时（秒）
```

---

### 4. TimeWindowCheck（时间窗口检查）

**功能**: 检查当前时间是否在指定窗口内

**配置参数**:
```yaml
windows:
  - name: "before_work"
    startTime: "00:00"
    endTime: "09:30"
    targetNode: "normal_clock_in"
  - name: "work_hours"
    startTime: "09:31"
    endTime: "17:30"
    targetNode: "check_late"
  - name: "after_work"
    startTime: "17:31"
    endTime: "23:59"
    targetNode: "terminate_after_work"
timezone: "Asia/Shanghai"
```

---

### 5. HolidayCheck（假期检查）

**功能**: 检查指定日期是否为工作日

**配置参数**:
```yaml
dateSource: "{{date}}"  # 或 "atom.timestamp"
holidayApiUrl: "https://unpkg.com/holiday-calendar@1.1.6/data/CN/"
saveIsWorkdayTo: "is_workday"
terminateIfHoliday: true  # 可选，假期则终止
```

**假期数据格式**（来自 holiday-calendar）:
```json
{
  "2025-01-01": {"name": "New Year", "name_cn": "元旦", "type": "public_holiday"},
  "2025-01-26": {"name": "Spring Festival Eve", "type": "transfer_workday"}
}
```

---

### 6. QueryDatabase（数据库查询）

**功能**: 执行 PostgreSQL 查询

**配置参数**:
```yaml
query: |
  SELECT COUNT(*) > 0 as on_leave
  FROM assignments_139
  WHERE slp_user_id = (SELECT id FROM forms_11 WHERE name = $1)
    AND slp_status NOT IN ('rejected', 'revoked')
    AND $2 BETWEEN "startDate" AND "endDate"
params:
  - "{{atom.person_name}}"
  - "{{date}}"
saveResultTo: "is_on_leave"
```

**安全约束**:
- 仅支持 SELECT 查询
- 参数化查询防止 SQL 注入
- 配置允许的表白名单

---

### 7. SkylarkFlow（Skylark 流程触发）

**功能**: 调用 Skylark API 创建流程

**配置参数**:
```yaml
flowId: 135
app: "smp.quanmate.com.cn"
userId: 97
data:
  name: "{{atom.person_name}}"
  time: "{{minute}}"
authHeader: "{{env.SKYLARK_JWT}}"
```

**调用流程**:
1. 获取字段映射 (GET /app/flows/:id)
2. 创建路由请求 (POST /flows/:id/journeys, operation: route)
3. 创建提议请求 (POST /flows/:id/journeys, operation: propose)

---

## 📅 实施计划

### Phase 4.1: 核心积木实现（优先）

- [ ] Switch 条件路由积木
- [ ] HttpRequest HTTP 请求积木
- [ ] CacheCheck / CacheSet 缓存操作积木
- [ ] TimeWindowCheck 时间窗口积木

### Phase 4.2: 外部服务积木

- [ ] HolidayCheck 假期检查积木
- [ ] QueryDatabase 数据库查询积木
- [ ] SkylarkFlow Skylark 流程触发积木

### Phase 4.3: 集成测试

- [ ] 创建门禁打卡逻辑图配置
- [ ] 模拟人脸识别数据触发测试
- [ ] 端到端流程验证

---

## 🔌 与 IoT/外部系统集成

人脸识别数据可通过两种方式接入：

1. **HTTP 直接触发**（推荐）：外部系统直接调用 lynxengine gRPC 接口
2. **MQTT 中转**：人脸识别系统发送 MQTT → IoT 引擎 → LynxGraph

推荐方案 1（HTTP 直接触发），因为人脸识别不是典型的 IoT 设备数据。

### REST API 触发方式

可考虑为 lynxmanager 添加 REST 接口，转发到 lynxengine gRPC：

```
POST /api/lynxgraph/trigger
{
  "infoAtomTypeId": "face_recognition",
  "data": {
    "Result": {
      "Description": "IMG001",
      "Tags": ["张三"],
      "Type": "face"
    },
    "TimeStamp": 1706150400000
  }
}
```

---

## 📚 参考文档

- `pkg/lynxgraph/DEPLOYMENT_PLAN.md` - LynxGraph 总体部署方案
- `pkg/lynxgraph/DEVELOPMENT.md` - LynxGraph 开发指南
- `D:\Documents\Code\access` - 原 Access 门禁打卡项目

---

**文档结束** | 最后更新: 2025-01-25
