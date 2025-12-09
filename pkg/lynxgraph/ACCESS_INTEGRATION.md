# 门禁打卡系统集成方案（Access 项目迁移）

## 概述

将独立的 Access 门禁打卡项目迁移到 LynxGraph 逻辑引擎，实现可视化配置打卡流程。

**当前状态**: ✅ 积木和边条件已实现，待集成测试

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
    name: "{{atom.person_name}}"
    time: "{{atom.time}}"
```

### 7. 迟到打卡（Flow 189）

```yaml
type: "skylark_journey_create"
config:
  app: "smp.quanmate.com.cn"
  flowId: 189
  userId: 97
  data:
    name: "{{atom.person_name}}"
    time: "{{atom.time}}"
```

---

## 边条件配置

通过边的 `condition` 字段实现分支，支持结构化 JSON。

**TimeWindowCheck 示例**（判断是否匹配 before_work 窗口）：
```json
{"logic":"and","conditions":[{"source":"context","path":"time_window.before_work","operator":"==","value":"true"}]}
```

**QueryDatabase 示例**（判断是否有请假记录）：
```json
{"logic":"and","conditions":[{"source":"context","path":"leave_check.data.on_leave","operator":"==","value":"true"}]}
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

## 待办

- [x] 实现 `dedup_check` 去重积木（基于图上下文）
- [x] 前端支持（字段下拉框、时间选择器）
- [x] `SkylarkJourneyCreate` 支持 `data` 字段映射
- [x] `QueryDatabase` 前端优化（动态参数列表 + PostgreSQL 语法校验）
- [ ] 创建逻辑图配置
- [ ] 端到端测试
