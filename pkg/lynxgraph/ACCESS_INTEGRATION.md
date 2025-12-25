# 门禁打卡系统集成方案（Access 项目迁移）

## 概述

将独立的 Access 门禁打卡项目迁移到 LynxGraph 逻辑引擎，实现可视化配置打卡流程。

**当前状态**: ✅ FilterBySet + ForEach + QueryDatabase 联调通过（2025-12-18）

---

## 已完成积木

| 积木 | 功能 | 实现文件 |
|-----|------|---------|
| HttpRequest | HTTP 请求（开门、音效） | `blocks/standard/httprequest/` |
| TimeWindowCheck | 时间窗口判断（上班前/后） | `blocks/standard/timewindow/` |
| Switch | 条件路由 | `blocks/standard/switch/` |
| HolidayCheck | 假期/工作日检查 | `blocks/standard/holidaycheck/` |
| QueryDatabase | 外部数据库查询 | `blocks/standard/querydatabase/` |
| SkylarkJourneyCreate | Skylark 流程触发 | `blocks/skylark/journeys/` |
| DedupCheck | 去重检查（每人每天一次） | `blocks/standard/dedupcheck/` |
| ForEach | 遍历数组执行子图 | `blocks/standard/foreach/` |
| GetDedupSet | 获取当前周期去重集合 | `blocks/standard/getdedupset/` |
| FilterBySet | 按集合过滤数组 | `blocks/standard/filterbyset/` |
| SetContainsCheck | 集合包含检查（已弃用，推荐使用 FilterBySet） | `blocks/standard/setcontainscheck/` |
| ClearDedupContext | 清空去重记录 | `blocks/standard/cleardedupcontext/` |
| LateTimeCalculate | 计算迟到分钟数 | `blocks/standard/latetimecalculate/` |

---

## 打卡业务逻辑

```
人脸识别 → 开门 → 去重检查 → 假期检查 → 时间窗口判断
                      │                  ├─ 上班前 → 正常打卡(Flow 135)
                      │                  └─ 上班后 → 请假/外勤检查
                      │                              ├─ 有 → 终止
                      │                              └─ 无 → 迟到打卡(Flow 189)
                      └─ 已处理 → 终止
```

### 去重策略

**关键点**：去重检查在**开门后**立即执行，无论后续判断结果如何（正常打卡、迟到、或因请假/外勤终止），都确保同一人同一天只触发一次打卡流程。开门动作始终执行，不受去重影响。

**实现方式**：基于图上下文（GraphContext），通过时间周期前缀实现自动重置。

```yaml
type: "dedup_check"
config:
  dedupFields:            # 去重字段
    - "person_name"
  resetMode: "daily"      # 重置周期：daily/hourly/none
  resetTime: "00:00"      # 每日重置时间（仅 daily 模式）
  timezone: "Asia/Shanghai"
  terminateIfExists: true # 已存在则终止流程（默认 true）
```

**前端展示**：
| 字段 | 控件 |
|------|------|
| 去重字段 | 多选下拉框（从入口节点订阅的信息原子字段中选择） |
| 重置周期 | 下拉框（每天/每小时/不重置） |
| 每日重置时间 | 时间选择器 |
| 已存在时终止 | 开关 |

**Key 生成示例**：
- daily 模式：`2025-12-05|person_name:张三`
- hourly 模式：`2025-12-05T10|person_name:张三`

> **注意**：去重标记在检查时**立即写入**图上下文，而非流程结束后。这样即使后续流程因请假/外勤而终止，也不会导致同一人当天被重复检查。

---

## 数据源配置

`service/lynxengine/etc/config.yaml`:

```yaml
ExternalDataSources:
  - Name: "skylark"
    Driver: "postgres"
    Host: "postgres"
    Port: 5432
    Database: "skylark"
    Username: "nexlyn"
    Password: "xxx"
```

---

## QueryDatabase 变量支持

QueryDatabase 积木的 `params` 参数支持以下变量语法：

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `{{atom.xxx}}` | InfoAtom Payload 中的字段 | `{{atom.name}}` → `王思未` |
| `{{timestamp}}` | InfoAtom 毫秒时间戳 | `1733630474000` |
| `{{date}}` | 日期字符串（Asia/Shanghai） | `2025-12-08` |
| `{{datetime}}` | 日期时间字符串 | `2025-12-08 12:01:14` |

> **注意**：时间相关字段（timestamp/date/datetime）从 InfoAtom 顶级 `Timestamp` 字段获取，不需要在信息原子的 `fields` 中定义。

---

## 积木配置示例

### 1. 开门 Webhook

```yaml
type: "http_request"
config:
  method: "POST"
  url: "http://192.168.6.95:8123/api/webhook/-O6Wb2..."
  ignoreError: true
```

### 2. 假期检查

```yaml
type: "holiday_check"
config:
  dateSource: "atom"
  holidayApiUrl: "https://unpkg.com/holiday-calendar@1.1.6/data/CN/"
  terminateIfHoliday: true
```

### 3. 时间窗口判断

```yaml
type: "time_window_check"
config:
  timeSource: "atom"
  windows:
    - name: "before_work"
      startTime: "00:00"
      endTime: "09:30"
    - name: "work_hours"
      startTime: "09:31"
      endTime: "17:30"
  resultKey: "time_window"
```

### 4. 查询请假记录

基于原 SQL（分两步查询）：
```sql
-- Step 1: 查用户ID（返回 int）
SELECT id FROM forms_11 WHERE name = $1
-- Step 2: 用 int 类型的 userId 查请假记录
SELECT ... FROM assignments_139 WHERE slp_user_id = $1 AND slp_status = 'finished'
  AND "startDate" <= $2 AND "endDate" >= $2
```

合并为单次查询的积木配置（注意：子查询需要类型转换）：

```yaml
type: "query_database"
config:
  dataSource: "skylark"
  query: |
    SELECT EXISTS (
      SELECT 1 FROM assignments_139
      WHERE slp_user_id = (SELECT id::integer FROM forms_11 WHERE name = $1)
        AND slp_status = 'finished'
        AND "startDate" <= $2 AND "endDate" >= $2
    ) AS on_leave
  params:
    - "{{atom.name}}"
    - "{{date}}"           # 使用内置变量，自动从 InfoAtom 时间戳转换
  saveResultTo: "leave_check"
  singleRow: true
```

### 5. 查询外勤记录

基于原 SQL（分两步查询）：
```sql
-- Step 1: 查用户ID（返回 int）
SELECT id FROM forms_11 WHERE name = $1
-- Step 2: 用 int 类型的 userId 查外勤记录
SELECT ... FROM assignments_99 WHERE slp_user_id = $1 AND slp_status = 'finished'
  AND "todayDate" = $2
```

积木配置：

```yaml
type: "query_database"
config:
  dataSource: "skylark"
  query: |
    SELECT EXISTS (
      SELECT 1 FROM assignments_99
      WHERE slp_user_id = (SELECT id::integer FROM forms_11 WHERE name = $1)
        AND slp_status = 'finished'
        AND "todayDate" = $2
    ) AS on_field_work
  params:
    - "{{atom.name}}"
    - "{{date}}"           # 使用内置变量，自动从 InfoAtom 时间戳转换
  saveResultTo: "fieldwork_check"
  singleRow: true
```

### 6. 正常打卡（Flow 135）

```yaml
type: "skylark_journey_create"
config:
  app: "smp.quanmate.com.cn"
  flowId: 135
  userId: 97
  data:                           # 可选，不配置则使用全部 InfoAtom 数据
    name: "{{atom.name}}"
    time: "{{atom.time}}"
```

### 7. 迟到时间计算

在时间窗口判断后、Skylark 流程触发前，使用此积木计算迟到分钟数。

```yaml
type: "late_time_calculate"
config:
  timeSource: "atom"           # 时间来源：atom（信息原子时间戳）或 now（当前时间）
  workStartTime: "09:30"       # 上班时间，HH:MM 格式
  timezone: "Asia/Shanghai"    # 时区
  saveResultTo: "late_time"    # 保存到 GraphContext 的键名
```

**输出格式**（保存到 GraphContext）：
```json
{
  "lateMinutes": 15,      // 迟到分钟数（整数，不迟到为0）
  "isLate": true,         // 是否迟到
  "checkTime": "09:45",   // 打卡时间（HH:MM）
  "workStartTime": "09:30" // 上班时间（HH:MM）
}
```

### 8. 迟到打卡（Flow 189）

```yaml
type: "skylark_journey_create"
config:
  app: "smp.quanmate.com.cn"
  flowId: 189
  userId: 97
  data:
    name: "{{atom.name}}"
    time: "{{context.late_time.lateMinutes}}"  # 从 GraphContext 获取迟到分钟数
```

### SkylarkJourneyCreate 变量语法

除了 `{{atom.xxx}}` 语法，现在还支持 `{{context.xxx.yyy}}` 从 GraphContext 获取数据：

| 语法 | 说明 | 示例 |
|------|------|------|
| `{{atom.xxx}}` | 信息原子 Payload 字段 | `{{atom.name}}` |
| `{{context.key.field}}` | 图上下文字段 | `{{context.late_time.lateMinutes}}` |

---

## 边条件配置

通过边的 `condition` 字段实现分支，支持结构化 JSON。

**TimeWindowCheck 示例**（判断是否匹配 before_work 窗口）：
```json
{"logic":"and","conditions":[{"source":"context","path":"time_window.before_work","operator":"==","value":"true"}]}
```

**QueryDatabase 示例**（判断是否有请假记录，单行模式下字段直接在顶层）：
```json
{"logic":"and","conditions":[{"source":"context","path":"leave_check.on_leave","operator":"==","value":"true"}]}
```

| 字段 | 说明 |
|-----|------|
| `source` | `context`（图上下文）或 `atom`（信息原子） |
| `path` | 变量路径，如 `time_window.before_work` |
| `operator` | `==` `!=` `>` `<` `>=` `<=` `contains` |
| `value` | 比较值，布尔类型用 `"true"` / `"false"` |
| `logic` | 多条件：`and` / `or` |

> 前端边条件表单会根据积木 OutputSchema 自动提示子字段和值类型。

---

## 信息原子定义

```yaml
name: "face_recognition"
fields:
  - name: "name"              # 人员姓名，用于 {{atom.name}}
    jsonPath: "$.Result.Tags[0]"
# 时间字段说明：
# - InfoAtom 顶级 Timestamp 字段从原始数据的 $.TimeStamp 自动提取
# - 积木中使用 {{date}} / {{datetime}} / {{timestamp}} 访问，无需在 fields 中定义
```

---

## 定时任务迁移

除了实时人脸识别打卡流程外，Access 项目还有两个定时任务需要迁移：

| 任务 | 触发时间 | 功能 |
|-----|---------|------|
| sendLateNotices | 每天 9:31 | 向未打卡且无请假/外勤的员工发送迟到通知 |
| processOffWorkLateClockIn | 每天 17:30 | 为未打卡且无请假/外勤的员工执行迟到打卡 |

### 核心架构：一个图 + 两个入口节点

**关键设计**：打卡流程和定时任务在**同一个逻辑图**中，通过**两个入口节点**分别触发，共享**同一个图上下文**。

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           打卡逻辑图（单一图）                                    │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌──────────────────────────────────┐   ┌────────────────────────────────────┐  │
│  │ 入口节点1（信息原子触发）          │   │ 入口节点2（定时触发）                │  │
│  │ 订阅: face_recognition           │   │ Schedule: 17:30 周一至周五          │  │
│  └───────────┬──────────────────────┘   └─────────────┬──────────────────────┘  │
│              │                                         │                         │
│              ▼                                         ▼                         │
│  ┌───────────────────────┐                ┌────────────────────────────────┐    │
│  │ 开门 (HttpRequest)    │                │ 假期检查 (HolidayCheck)        │    │
│  └───────────┬───────────┘                └─────────────┬──────────────────┘    │
│              │                                          │                        │
│              ▼                                          ▼                        │
│  ┌───────────────────────┐                ┌────────────────────────────────┐    │
│  │ 去重检查 (DedupCheck) │◄───────────────┤ 获取已打卡集合 (GetDedupSet)   │    │
│  │ 记录: person_name     │   共享图上下文  │ 关联同一个 DedupCheck 节点      │    │
│  └───────────┬───────────┘                └─────────────┬──────────────────┘    │
│              │                                          │                        │
│              ▼                                          ▼                        │
│        ... 后续打卡逻辑 ...                   FilterBySet(排除已打卡员工)         │
│                                                         │                        │
│                                                         ▼                        │
│                                               ForEach → 子图处理未打卡员工       │
│                                                         │                        │
│                                                         ▼                        │
│                                            ClearDedupContext (清空去重记录)      │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**优势**：
1. **共享图上下文**：DedupCheck 写入的去重记录，GetDedupSet 可以直接读取
2. **配置联动**：GetDedupSet、ClearDedupContext 可以关联 DedupCheck 节点，自动使用相同的去重配置
3. **逻辑内聚**：整个打卡业务（实时打卡 + 定时处理）在一个图里，更易理解和维护

### 定时任务业务流程

```
定时触发 → 假期检查 → 查询所有员工 → 获取已打卡集合 → FilterBySet(排除已打卡)
    → 遍历员工(ForEach)
        → 检查请假 → 有请假则跳过
        → 检查外勤 → 有外勤则跳过
        → 触发 Skylark 流程（迟到通知/迟到打卡）
    → [仅下班任务] 清空当日去重记录
```

> **注意**：使用 FilterBySet 在 ForEach 之前过滤已打卡员工，而非在子图中使用 SetContainsCheck。这是因为子图无法访问主图的 GraphContext。

### 已完成组件

| 组件 | 位置 | 说明 |
|-----|------|------|
| ScheduleManager | `internal/scheduler/` | 定时调度管理器（含分布式锁） |
| ScheduleBlock | `internal/blocks/standard/schedule/` | 定时触发积木（在画布上可视化配置） |
| ScheduleConfig | `core/schedule.go` | 定时配置结构体和辅助函数 |
| DispatchScheduled | `internal/dispatcher/handler.go` | 定时触发分发（只触发 schedule 入口节点） |
| RegisterAllSchedules | `internal/graph/handler.go` | 解决初始化顺序问题的延迟注册 |

### ForEach 积木（信息原子驱动模式）

**关键设计**：子图也是图，需要信息原子来驱动。ForEach 通过创建信息原子触发订阅了该类型的子图。

```
ForEach 积木 → 遍历数组 → 为每个元素创建信息原子 → Dispatcher 分发 → 子图入口节点匹配 → 子图执行
```

**使用步骤**：
1. 用户创建迭代信息原子类型（如 `foreach.late_employee`）
2. 创建子图，入口节点订阅该信息原子类型
3. 主图中 ForEach 积木配置使用该信息原子类型
4. 执行时 ForEach 为每个元素创建信息原子并分发

#### ForEach 配置

```yaml
type: "foreach"
config:
  sourceType: "context"                    # "context" | "atom"
  sourcePath: "employees.data"             # 数组路径
  iterationInfoAtomTypeId: "xxx-uuid-xxx"  # 迭代信息原子类型ID（前端选择）
  itemKey: "employee"                      # Payload 中元素的 key，默认 "item"
  indexKey: "idx"                          # Payload 中索引的 key，默认 "index"
  continueOnError: true                    # 单元素失败是否继续
```

**ForEach 提供的原始 Payload**（用于 DataFormat 字段提取）：
```json
{
  "employee": { "name": "张三", "id": 1 },  // 当前元素（使用 itemKey 配置）
  "idx": 0,                                 // 当前索引（使用 indexKey 配置）
  "total": 10,                              // 数组总长度
  "parentGraphKey": "xxx",                  // 父图 GraphKey
  "parentNodeId": "yyy"                     // 父图 ForEach 节点 ID
}
```

**DataFormat 字段提取**：子图收到的信息原子 Payload 根据迭代信息原子类型的 DataFormat 配置提取。

迭代信息原子类型 DataFormat 配置示例：
```json
{
  "fields": [
    {"fieldKey": "name", "fieldPath": "employee.name", "fieldType": "string"},
    {"fieldKey": "currentIndex", "fieldPath": "idx", "fieldType": "int"}
  ]
}
```

子图实际收到的 Payload：
```json
{
  "name": "张三",
  "currentIndex": 0
}
```

> **注意**：如需 `parentGraphKey` 等字段，需在 DataFormat 中显式配置。

**子图访问迭代变量**（从信息原子 Payload 获取，根据 DataFormat 配置）：
- `{{atom.name}}` - 当前员工姓名（对应 fieldKey）
- `{{atom.currentIndex}}` - 当前索引（对应 fieldKey）

#### GetDedupSet 配置

**方式一：关联 DedupCheck 节点（推荐）**

```yaml
type: "get_dedup_set"
config:
  sourceNodeId: "node_dedup_check_xxx"  # 关联图中的 DedupCheck 节点
  saveResultTo: "checked_employees"
  # dedupFields/resetMode/resetTime/timezone 从关联节点自动获取
```

**方式二：手动配置（兼容旧方式）**

```yaml
type: "get_dedup_set"
config:
  dedupFields:                    # 与 DedupCheck 一致
    - "person_name"
  resetMode: "daily"
  resetTime: "00:00"
  timezone: "Asia/Shanghai"
  saveResultTo: "checked_employees"
```

**输出**（保存到 GraphContext）：
```json
{
  "keys": ["2025-12-16|name:张三", "2025-12-16|name:李四"],
  "records": [{"name": "张三"}, {"name": "李四"}],
  "count": 2,
  "timePeriod": "2025-12-16",
  "queriedAt": "2025-12-16T17:30:00+08:00"
}
```

> **注意**：`records` 中的字段名与 `dedupFields` 配置一致（如 `dedupFields: ["name"]` 则输出 `{"name": "张三"}`）。

#### FilterBySet 配置

在 ForEach 之前过滤数组，排除已存在于集合中的元素。

```yaml
type: "filter_by_set"
config:
  sourceType: "context"
  sourcePath: "employee_list.data"     # QueryDatabase 查询结果的数组路径
  setContextKey: "checked_employees"   # GetDedupSet 的 saveResultTo
  matchField: "name"                   # 源数组元素中的匹配字段
  setMatchField: "name"                # 集合 records 中的匹配字段（可选，默认同 matchField）
  saveResultTo: "unchecked_employees"  # 过滤后的未打卡员工
  invertFilter: false                  # false=排除存在的，true=保留存在的
```

**字段匹配说明**：
- `matchField`：源数组（如 QueryDatabase 结果）中用于匹配的字段
- `setMatchField`：GetDedupSet 输出的 `records` 中用于匹配的字段
- **关键**：`setMatchField` 必须与 DedupCheck 的 `dedupFields` 配置一致（如都用 `name`）

**输出格式**（保存到 GraphContext）：
```json
{
  "data": [...],      // 过滤后的数组
  "count": 5,         // 过滤后的数量
  "filteredCount": 3  // 被过滤掉的数量
}
```

#### SetContainsCheck 配置（已弃用）

> **注意**：SetContainsCheck 设计用于子图中检查，但由于子图无法访问主图的 GraphContext，此积木无法正常工作。推荐使用 FilterBySet 在 ForEach 之前过滤。

```yaml
type: "set_contains_check"
config:
  setContextKey: "checked_employees"     # GetDedupSet 的 saveResultTo
  checkField: "person_name"              # 去重字段名（与 GetDedupSet.dedupFields 一致）
  valueExpr: "{{atom.name}}"             # 从信息原子获取检查值
  terminateIfExists: true                # 已存在则终止（默认 true）
```

#### ClearDedupContext 配置

清空当前时间周期的去重记录，放在 ForEach 之后执行。

**方式一：关联 DedupCheck 节点（推荐）**

```yaml
type: "clear_dedup_context"
config:
  sourceNodeId: "node_dedup_check_xxx"  # 关联图中的 DedupCheck 节点
  # dedupFields/resetMode/resetTime/timezone 从关联节点自动获取
```

**方式二：手动配置（兼容旧方式）**

```yaml
type: "clear_dedup_context"
config:
  dedupFields:              # 与 DedupCheck 一致
    - "person_name"
  resetMode: "daily"
  resetTime: "00:00"
  timezone: "Asia/Shanghai"
```

### 打卡逻辑图（完整配置）

一个图包含两个入口节点，分别处理实时打卡和定时任务：

**入口节点1（信息原子触发）**：实时人脸识别打卡
```
入口节点(订阅 face_recognition)
    → 开门(HttpRequest)
    → 去重检查(DedupCheck, dedupFields=["person_name"])
    → 假期检查(HolidayCheck, terminateIfHoliday=true)
    → 时间窗口判断(TimeWindowCheck)
    → [边条件分支] ...后续打卡流程
```

**入口节点2（定时触发）**：下班迟到打卡
```
定时入口(Schedule, cronExpr="30 17 * * 1-5")
    → 假期检查(HolidayCheck, terminateIfHoliday=true)
    → 查询所有员工(QueryDatabase → employees)
    → 获取已打卡集合(GetDedupSet, sourceNodeId=去重检查节点ID)
    → 过滤已打卡员工(FilterBySet, setContextKey=checked_employees, saveResultTo=unchecked_employees)
    → 遍历未打卡员工(ForEach, sourcePath=unchecked_employees.data, iterationInfoAtomTypeId=late_clock_in_iteration)
    → 清空去重记录(ClearDedupContext, sourceNodeId=去重检查节点ID)
```

**ForEach 子图**（订阅 `late_clock_in_iteration` 信息原子类型）：
```
入口节点(订阅 late_clock_in_iteration)
    → 检查请假(QueryDatabase → leave_check)
    → [边条件: leave_check.data.on_leave == false]
    → 检查外勤(QueryDatabase → fieldwork_check)
    → [边条件: fieldwork_check.data.on_field_work == false]
    → 迟到打卡(SkylarkJourneyCreate, flowId=189)
```

> **注意**：子图中不再需要 SetContainsCheck，因为已打卡员工在 FilterBySet 阶段就被排除了。

**定时积木配置**：

```yaml
type: "schedule"
config:
  cronExpr: "30 17 * * 1-5"     # 周一至周五 17:30
  timezone: "Asia/Shanghai"
  initialPayload:
    taskType: "late_clock_in"
```

### 关键配置关联

| 配置项 | 关联项 | 说明 |
|--------|--------|------|
| GetDedupSet.`sourceNodeId` | DedupCheck 节点 ID | 自动使用 DedupCheck 的去重配置 |
| ClearDedupContext.`sourceNodeId` | DedupCheck 节点 ID | 自动使用 DedupCheck 的去重配置 |
| GetDedupSet.`saveResultTo` | FilterBySet.`setContextKey` | 去重集合用于过滤 |
| DedupCheck.`dedupFields` | FilterBySet.`setMatchField` | **必须一致**（如都用 `name`） |
| FilterBySet.`saveResultTo` | ForEach.`sourcePath` | 过滤后的数组用于遍历（如 `unchecked_employees.data`） |
| ForEach.`iterationInfoAtomTypeId` | 子图入口节点.`subscribedInfoAtomTypes` | 信息原子类型绑定 |

---

## 待办

- [x] 实现 `dedup_check` 去重积木（基于图上下文）
- [x] 前端支持（字段下拉框、时间选择器）
- [x] `SkylarkJourneyCreate` 支持 `data` 字段映射
- [x] `QueryDatabase` 前端优化（动态参数列表 + PostgreSQL 语法校验）
- [ ] 创建逻辑图配置
- [ ] 端到端测试

### 定时任务迁移

- [x] 实现 `ScheduleManager` 定时调度管理器（含分布式锁）
- [x] 实现 `ScheduleBlock` 定时触发积木（在画布上可视化配置）
- [x] 实现 `DispatchScheduled` 定时分发（只触发 schedule 入口节点）
- [x] 实现 `RegisterAllSchedules` 解决初始化顺序问题
- [x] 实现 `ForEach` 循环积木（信息原子驱动模式）
- [x] 实现 `GetDedupSet` 获取去重集合积木
- [x] ForEach + QueryDatabase + 子图联调验证通过（2025-12-17）
- [x] 前端 `infoAtomTypeSelector` 支持（ForEach 迭代信息原子类型下拉选择）
- [x] ForEach DataFormat 字段提取（子图收到的信息原子符合 DataFormat 定义）
- [x] 实现 `SetContainsCheck` 集合包含检查积木（2025-12-17，已弃用）
- [x] 实现 `ClearDedupContext` 清空去重记录积木（2025-12-17）
- [x] 实现去重积木配置联动（GetDedupSet/ClearDedupContext 可关联 DedupCheck 节点，2025-12-17）
- [x] 前端 `dedupCheckNodeSelector` 和 `contextKeys` 支持（2025-12-17）
- [x] 实现 `FilterBySet` 按集合过滤积木（2025-12-18，替代 SetContainsCheck）
- [x] 实现 `LateTimeCalculate` 迟到时间计算积木（2025-12-23）
- [x] 扩展 `SkylarkJourneyCreate` 支持 `{{context.xxx}}` 变量语法（2025-12-23）
- [ ] 创建打卡逻辑图（一个图两个入口节点）
- [ ] 定时任务端到端测试
