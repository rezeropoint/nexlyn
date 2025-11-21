# LynxGraph 前端页面设计规范文档

## 📋 文档概述

本文档描述 LynxGraph 逻辑引擎前端管理界面的架构设计、功能规划和实施进度。

**目标**：
- 提供逻辑图配置和可视化编辑界面
- 支持运行时数据监控和统计分析
- 实现编辑-监控-调试的完整工作流
- 遵循 Ant Design 设计规范和项目编码规范

**最后更新**: 2025-10-26

---

## 🏗️ 整体架构设计

### 核心设计决策

采用 **卡片+表格双视图列表页 + 独立详情页（Tab模式）** 的架构。

**设计理念**：
1. **场景区分**：逻辑图配置不同于普通CRUD，需要"配置+监控"混合场景支持
2. **视图切换**：列表页提供卡片视图（浏览监控）和表格视图（批量管理）双模式
3. **编辑-监控一体化**：详情页Tab模式支持快速切换编辑和监控功能
4. **渐进式增强**：优先实现核心编辑功能，监控统计功能后续对接真实API

### 页面架构

```
┌──────────────────────────────────────────────────────┐
│  列表页 (/logic-engine/graph-config)                 │
│  ┌────────────────────────────────────────────────┐  │
│  │ 视图切换：[卡片视图 ✓] / [表格视图]           │  │
│  └────────────────────────────────────────────────┘  │
│                                                       │
│  卡片视图（默认）：                                   │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐                │
│  │ 图标 │ │ 图标 │ │ 图标 │ │ 图标 │                │
│  │ 图1  │ │ 图2  │ │ 图3  │ │ 图4  │                │
│  │ 统计 │ │ 统计 │ │ 统计 │ │ 统计 │                │
│  └──────┘ └──────┘ └──────┘ └──────┘                │
│                                                       │
│  表格视图：ProTable（搜索、筛选、批量操作）          │
│                                                       │
│  点击卡片/图名称 → 跳转详情页                        │
└──────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────┐
│  详情页 (/logic-engine/graph-config/:id)             │
│  ┌────────────────────────────────────────────────┐  │
│  │ Tab1: 基本信息（编辑元数据）                   │  │
│  │ Tab2: 可视化编辑器（AntV X6图形编辑）          │  │
│  │ Tab3: 运行时数据（执行日志+图上下文）          │  │
│  │ Tab4: Webhooks配置（外部集成）                 │  │
│  │ Tab5: 统计分析（执行趋势+性能指标）            │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

---

## 🎨 路由设计

### 路由配置

```typescript
{
  path: '/logic-engine',
  name: '逻辑引擎',
  icon: 'ApartmentOutlined',
  routes: [
    {
      path: '/logic-engine/graph-config',
      name: '逻辑图配置',
      component: './LogicEngine/GraphConfig',
      exact: true,
    },
    {
      path: '/logic-engine/graph-config/:id',
      name: '逻辑图详情',
      component: './LogicEngine/GraphConfig/Detail',
      hideInMenu: true,
    },
  ],
}
```

### URL参数

- `:id` - 逻辑图ID（UUID格式）
- 查询参数 `?tab=basic|visual|runtime|webhooks|stats` - 指定默认Tab
- 查询参数 `?view=card|table` - 指定列表页默认视图模式（默认: card）

---

## 📂 目录结构

```
frontend/src/pages/LogicEngine/GraphConfig/
├── index.tsx                          # 列表页主入口
├── types.ts                           # 类型定义
├── constants.ts                       # 常量定义
├── iconConfig.tsx                     # 图标配置（32个预定义图标）
│
├── hooks/
│   └── useGraphConfig.ts              # 列表数据管理Hook
│
├── components/
│   ├── GraphConfigTable.tsx           # 表格视图组件
│   ├── GraphConfigTableActions.tsx    # 表格操作列
│   ├── GraphConfigCardView.tsx        # 卡片视图容器
│   ├── GraphConfigCard.tsx            # 单个卡片组件
│   ├── GraphConfigCard.less           # 卡片样式
│   ├── IconSelector.tsx               # 图标选择器
│   ├── IconSelector.less              # 图标选择器样式
│   ├── CreateGraphConfigForm.tsx      # 创建表单
│   ├── EditGraphConfigForm.tsx        # 编辑表单
│   ├── NodeListEditor.tsx             # 节点列表编辑器
│   └── EdgeListEditor.tsx             # 边列表编辑器
│
└── Detail/                            # 详情页目录
    ├── index.tsx                      # 详情页主入口（Tab容器）
    ├── hooks/
    │   └── useGraphDetail.ts          # 详情数据管理Hook
    │
    └── components/
        ├── BasicInfoTab/              # Tab1: 基本信息
        ├── VisualEditorTab/           # Tab2: 可视化编辑器
        │   ├── BlockPanel.tsx         #   - 左侧逻辑块面板
        │   ├── GraphCanvas.tsx        #   - 中间X6画布
        │   ├── PropertyPanel.tsx      #   - 右侧属性面板
        │   └── Toolbar.tsx            #   - 顶部工具栏
        ├── RuntimeDataTab/            # Tab3: 运行时数据
        ├── WebhooksTab/               # Tab4: Webhooks配置
        └── StatsTab/                  # Tab5: 统计分析
```

---

## 🔧 技术栈选型

| 技术 | 版本 | 用途 |
|------|------|------|
| **@antv/x6** + 插件 | ^2.x | 图形编辑器（Snapline/History/Keyboard/Selection/Clipboard） |
| **dagre** | latest | 自动布局算法 |
| **react-icons** | ^5.x | 图标库（30,000+图标） |
| **nanoid** | latest | 生成节点/边ID |
| **@ant-design/charts** | 已安装 | 图表组件 |
| **@ant-design/pro-components** | 已安装 | 高级表单组件 |

---

## 📄 列表页详细设计

### 功能特性

**✅ 已实现功能**：
- 卡片/表格双视图切换（Segmented组件）
- 默认卡片视图展示
- 图标系统集成（32个预定义图标，8个类别）
- 搜索/筛选：按名称、版本、标签、启用状态
- 快速操作：查看详情、复制、删除
- 点击图名称/卡片跳转详情页
- 面包屑导航修复（使用href）

**卡片视图特性**：
- 展示图标（支持自定义或默认图标）
- 显示运行统计（节点数、执行次数、成功率）
- 视觉化展示图的状态和版本
- 响应式网格布局，支持主题切换

**表格视图特性**：
- 高密度信息展示
- 支持多列排序和筛选
- 批量选择和操作

### 图标系统设计

**预定义图标分类**（共32个）：
- **AI/智能**（4个）：机器人、脑部、芯片、网络
- **物联网**（4个）：传感器、温度计、摄像头、设备
- **数据处理**（4个）：数据库、图表、过滤器、转换
- **流程/工作流**（4个）：流程图、节点、自动化、GitLab
- **监控/告警**（4个）：仪表盘、警报、通知、活动
- **安全**（4个）：安全防护、认证、加密、盾牌
- **集成/连接**（4个）：API、Webhook、插件、集成
- **媒体**（4个）：视频、音频、流媒体、播放

**图标选择器功能**：
- 按类别浏览（Tabs标签页）
- 搜索图标（支持中文标签和英文名称）
- 图标预览（32px大小，带颜色）
- 主题色自动适配

### API调用

| API | 方法 | 说明 | 状态 |
|-----|------|------|------|
| `/api/lynxmanager/graph/list` | GET | 逻辑图列表查询 | ✅ 已实现 |
| `/api/lynxmanager/graph/delete/:id` | DELETE | 删除逻辑图 | ✅ 已实现 |
| `/api/lynxmanager/graph/create` | POST | 创建逻辑图（含icon字段） | ⏳ 待更新 |
| `/api/lynxmanager/graph/update/:id` | PUT | 更新逻辑图（含icon字段） | ⏳ 待更新 |

---

## 📄 详情页设计

### Tab 1: 基本信息

**功能定位**：编辑逻辑图元数据

**功能**：
- Descriptions组件展示只读信息
- 编辑按钮打开ModalForm
- 可编辑字段：名称、版本、描述、标签、图标、启用状态

**状态**：✅ 已实现

---

### Tab 2: 可视化编辑器

**功能定位**：图形化编辑节点和边

**布局**：
- 顶部工具栏：保存、撤销、重做、自动布局
- 左侧面板（220px）：逻辑块列表（按类别分组）
- 中间画布（flex 1）：AntV X6画布
- 右侧面板（320px）：节点/边属性编辑器

**状态**：✅ 已实现

---

### Tab 3: 运行时数据

**功能定位**：查看逻辑图执行情况

**数据展示**：
- 执行日志表格
- 信息原子列表
- 图上下文查看器（react-json-view）

**状态**：✅ UI已实现（Mock数据），⏳ 待对接真实API

---

### Tab 4: Webhooks配置

**功能定位**：配置外部系统集成

**状态**：✅ UI已实现（Mock数据），⏳ 待对接真实API

---

### Tab 5: 统计分析

**功能定位**：可视化展示执行统计

**图表类型**：
- 统计卡片：执行总数、成功次数、失败次数、成功率
- 执行趋势图：24小时折线图
- 性能指标图：平均耗时、P95/P99耗时
- 节点执行统计：各节点执行次数分布

**状态**：✅ UI已实现（Mock数据），⏳ 待对接真实API

---

## 🌐 核心数据模型

### GraphConfigMetadata（扩展）

```typescript
interface GraphConfigMetadata {
  id: string;
  tenantId: string;
  name: string;
  version: string;
  description?: string;
  tags?: string[];
  isEnabled: boolean;

  // 🆕 图标字段
  icon?: string;           // 图标名称（如 'FaRobot', 'MdSensors'）
  iconColor?: string;      // 图标颜色（可选，默认使用主题色）

  // 运行统计（后续实现）
  nodeCount?: number;      // 节点数量
  executionCount?: number; // 总执行次数
  successRate?: number;    // 成功率（%）
  lastExecutedAt?: number; // 最后执行时间

  // 审计字段
  createdAt?: number;
  updatedAt?: number;
  createdBy?: string;
  updatedBy?: string;
}
```

### NodeItem 和 EdgeItem

```typescript
interface NodeItem {
  key?: string;            // UI用，不提交到后端
  id: string;
  blockType: string;
  blockVersion: string;
  isEntryPoint: boolean;
  subscribedInfoAtomTypeIDs?: string[];
  subscribedSource?: string;
  subscribedLabels?: string[];
  blockConfig?: string;    // JSON字符串
  x?: number;              // 画布位置
  y?: number;
}

interface EdgeItem {
  key?: string;            // UI用，不提交到后端
  id: string;
  sourceID: string;
  targetID: string;
  condition?: string;
}
```

---

## 🎨 样式规范

**核心原则**：
- 禁止内联样式，使用Less模块
- 使用CSS变量适配主题：`var(--ant-color-*)`
- Modal/Drawer使用 `destroyOnHidden`

**编辑器布局**：
- 容器高度：`calc(100vh - 240px)`
- 左侧面板：220px（最小180px）
- 中间画布：flex: 1
- 右侧面板：320px（最小280px，可折叠）

---

## 📋 开发清单

### ✅ Phase 1: 核心功能（已完成）

- [x] 列表页（表格视图）
- [x] 路由配置
- [x] 详情页Tab容器框架
- [x] Tab1: 基本信息展示和编辑
- [x] Tab2: 可视化编辑器基础功能
- [x] Tab3: 运行时数据UI（Mock数据）
- [x] Tab4: Webhooks配置UI（Mock数据）
- [x] Tab5: 统计分析UI（Mock数据）

### ✅ Phase 2: 交互优化与图标系统（已完成）

- [x] 列表页：实现卡片视图
- [x] 列表页：视图切换功能（Segmented）
- [x] 列表页：优化操作列（查看详情、复制、删除）
- [x] 列表页：移除行点击事件，改为点击名称链接
- [x] 详情页：修复面包屑点击（使用href）
- [x] 图标系统：集成react-icons（32个预定义图标）
- [x] 图标选择器：搜索、分类浏览、主题适配
- [x] 创建/编辑表单：集成图标选择器
- [x] 卡片显示：图标+标题+版本在同一行
- [x] TypeScript/ESLint检查通过

### ✅ Phase 3: 后端API对接（部分完成）

- [x] 后端API：支持icon/iconColor字段的创建和更新（PostgreSQL存储）
- [x] 后端API：支持节点X/Y坐标字段的存储和查询（MongoDB存储）
- [x] 前端表单：集成IconSelector组件（创建/编辑表单）
- [x] 前端类型：完善CreateGraphConfigRequest/UpdateGraphConfigRequest类型定义
- [ ] 后端API：List接口返回运行统计字段（nodeCount、executionCount、successRate）
- [ ] Tab3: 对接执行日志API
- [ ] Tab3: 对接信息原子API
- [ ] Tab3: 对接图上下文API
- [ ] Tab5: 对接统计分析API

### ⏳ Phase 4: 集成功能对接（待实现）

- [ ] Tab4: 对接Webhook CRUD API
- [ ] Tab4: 实现Webhook测试功能

### ✅ Phase 5: 编辑器增强（已完成）

- [x] Tab2: 拖拽添加节点（从逻辑块面板拖拽）
- [x] Tab2: 可视化连线编辑（点击端口连接）
- [x] Tab2: 复制粘贴节点（Ctrl+C/V）
- [x] Tab2: 快捷键支持（Delete/Ctrl+Z/Ctrl+Y）
- [x] Tab2: 撤销/重做功能（History插件）
- [x] Tab2: 自动布局算法（dagre）
- [x] Tab2: 对齐线辅助（Snapline插件）
- [x] 三栏可调整布局（Splitter组件）
- [x] 属性面板实时编辑
- [x] 工具栏操作（保存/缩放/居中/清空）

### ⏳ Phase 6: 高级功能（待实现）

- [ ] Tab2: 小地图导航（MiniMap插件）
- [ ] 导出/导入逻辑图JSON
- [ ] 实时数据刷新（WebSocket）
- [ ] 节点验证（检查环/入口节点）

---

## 📚 参考资源

**官方文档**：[AntV X6](https://x6.antv.vision/zh) | [react-icons](https://react-icons.github.io/react-icons/) | [ProComponents](https://procomponents.ant.design/)

**项目文档**：`frontend/前端编码规范与最佳实践.md` | `restful/lynxmanager/api/lynxmanager.api` | `pkg/lynxgraph/DEVELOPMENT.md`

---

## 📊 实施进度

**最后更新**: 2025-10-27

### ✅ Phase 1-2-3-5 已完成（核心稳定）

**核心功能**：
- ✅ 列表页（卡片/表格双视图、图标系统）
- ✅ Tab1: 基本信息编辑（含图标选择）
- ✅ Tab2: 完整可视化编辑器（拖拽、连线、属性编辑、坐标持久化）
- ✅ Tab3-5: UI框架（Mock数据）
- ✅ 图标字段完整支持（前端→后端→数据库）
- ✅ 节点坐标完整持久化（前端画布→MongoDB存储）

**编辑器特性**：
- ✅ 三栏布局（逻辑积木面板/画布/属性面板）
- ✅ 拖拽添加节点（原生HTML5 Drop API）、节点坐标持久化（x/y字段）
- ✅ 撤销/重做（初始加载不计入历史）、自动布局、快捷键
- ✅ **稳定性**：已修复点击节点消失、画布闪烁、选中特效不更新等核心bug

**数据持久化**：
- ✅ PostgreSQL：图元数据（icon、iconColor等）
- ✅ MongoDB：节点详情（X、Y坐标等）

**技术栈**：
- ✅ AntV X6 + 5个插件（Snapline/History/Keyboard/Selection/Clipboard）
- ✅ dagre自动布局、nanoid生成ID、TypeScript/ESLint通过

### ⏳ Phase 4-6 待实现

- ⏳ 运行统计数据对接、监控API、Webhook功能
- ⏳ 小地图导航、导出/导入、实时刷新

---

## 🔄 设计变更记录

### v4.7 (2025-11-03)

**属性面板背景与滚动修复**：

1. 将 `PropertyPanel` 样式切换为 CSS Modules（`PropertyPanel.module.less`），保留 `:global` 选择器精准覆盖 AntD 内部结构，避免全局样式污染。
2. 组件侧改为按模块化 className 引用（`styles.propertyPanelCard` 等），确保构建后的散列类名与样式一致。
3. 修复后属性面板 `ant-card-body` 背景与滚动行为正确继承主题变量，MCP 自动化验证背景为 `rgb(255,255,255)` 且 `overflow-y: auto`，滚动高度与内容一致。
4. 富余样式（Tag 间距、空态提示、JSON 编辑器字体）一并迁移至模块文件，遵循《前端编码规范与最佳实践》关于 CSS Modules 与主题变量的要求。

### v4.6 (2025-10-27)

**新建逻辑图流程优化**：

1. **创建表单简化**：
   - 移除节点和边的配置表单（NodeListEditor、EdgeListEditor组件）
   - 仅保留基本信息字段：名称、版本、图标、描述、标签、启用状态
   - 表单宽度从 1200px 缩小为 720px，提升用户体验
   - 创建时提交空节点和边数组（`nodes: []`, `edges: []`）

2. **API字段调整**：
   - 修改 `CreateGraphConfigRequest` 的 `nodes` 和 `edges` 字段为可选（添加 `optional` 标签）
   - API定义更新：`Nodes []NodeConfig \`json:"nodes,optional"\`` 和 `Edges []EdgeConfig \`json:"edges,optional"\``
   - types.go 自动同步更新（通过 goctl 生成）

3. **设计理念**：
   - **两阶段工作流**：创建阶段填写元数据 → 详情页编辑阶段配置节点和边
   - **职责分离**：列表页创建表单专注基本信息，Tab 2 可视化编辑器负责图结构编辑
   - **用户友好**：避免在新建时填写复杂的节点和边配置，降低使用门槛

4. **影响范围**：
   - 前端：`CreateGraphConfigForm.tsx` 组件精简（移除 NodeListEditor、EdgeListEditor 导入和表单项）
   - 后端：`restful/lynxmanager/api/lynxmanager.api` 更新（nodes/edges 字段改为可选）
   - 类型：`restful/lynxmanager/internal/types/types.go` 自动同步

### v4.5 (2025-10-27)

**后端数据持久化完整支持**：

1. **图标字段完整存储**：
   - 后端：添加PostgreSQL表字段（icon、icon_color）+ 数据库模型 + SQL语句更新
   - API：更新lynxmanager.api定义，包含icon/iconColor字段
   - 类型：前端types.ts添加CreateGraphConfigRequest/UpdateGraphConfigRequest的icon字段
   - 验证：后端编译通过、前端TypeScript检查通过

2. **节点坐标持久化支持**：
   - 核心模型：pkg/lynxgraph/core/grap.go NodeConfig添加X/Y字段
   - 数据流：前端画布坐标→API传输→MongoDB存储→查询返回→画布渲染
   - 转换层：restful/lynxmanager logic/converter层完整支持坐标字段传递

3. **数据存储架构**：
   - PostgreSQL：图元数据（名称、版本、标签、icon、icon_color等）
   - MongoDB：节点详情（id、blockType、X、Y等）+ 边详情

### v4.4 (2025-10-27)

**滚动条和框选冲突修复 + 主题适配完善 + Splitter 稳定性修复**：

1. **彻底移除 Scroller 插件**：Scroller 插件会创建自己的滚动条容器（`.x6-graph-scroller`），即使用CSS隐藏，其 `pannable: true` 仍会与 Selection 插件的框选功能冲突，导致框选时同时触发画布平移

2. **改用原生 panning 配置**：使用 X6 原生的 `panning` 配置，通过 `modifiers: 'shift'` 区分框选和平移操作，彻底解决冲突

3. **修复画布容器滚动条**（多层防护）：
   - **容器样式**：`.graph-canvas` 添加 `overflow: hidden` + `box-sizing: border-box` + `max-width/max-height: 100%`，防止任何溢出
   - **内部容器**：`.x6-graph` 强制设置 `overflow: hidden !important`，防止 X6 内部容器产生滚动条
   - **尺寸计算**：使用 `clientWidth/clientHeight`（已排除边框），精确匹配容器内容区域

4. **全局滚动条主题适配**（新增）：
   - **统一样式管理**：在 `VisualEditorTab/index.less` 添加全局滚动条样式（通配符选择器 `*`）
   - **主题变量**：使用 CSS 变量确保滚动条自动适配亮色/暗色主题
   - **覆盖范围**：所有子面板（逻辑积木面板、画布、属性面板）的滚动条统一使用主题适配样式
   - **样式精简**：移除 `BlockPanel.less` 和 `PropertyPanel.less` 中重复的滚动条定义，继承父容器样式

5. **修复 Splitter 阈值交换问题**（新增）：
   - **问题根源**：混用像素值（`defaultSize="220px"`）和百分比值（`max="30%"`）导致拖动时约束计算错误
   - **解决方案**：统一使用像素值（数字类型），避免单位混用
   - **配置优化**：
     - 左侧面板：`defaultSize={260} min={220} max={450}`
     - 中间画布：`min={480}`（自动填充剩余空间）
     - 右侧面板：`defaultSize={320} min={280} max={450}`
   - **稳定性保证**：所有约束使用一致的单位，消除阈值交换 bug

6. **操作方式优化**：
   - **框选**：直接拖拽空白区域进行框选（Selection 插件的 `rubberband: true`）
   - **平移**：按住 Shift 键拖拽空白区域平移画布（原生 `panning` 配置）
   - **缩放**：Ctrl + 滚轮缩放（`mousewheel` 配置）

7. **纯净画布体验**：彻底消除所有滚动条（包括Scroller插件的、浏览器原生的、X6内部容器的），所有操作通过鼠标和快捷键完成；所有滚动条完美适配主题切换

**技术栈变更**：
- 移除依赖：`@antv/x6-plugin-scroller`（不再需要）

**样式架构改进**：
- ✅ 父容器统一管理滚动条样式（`index.less`）
- ✅ 子组件继承父容器样式（移除重复定义）
- ✅ 使用 CSS 变量确保主题适配

**Splitter 配置规范**：
- ✅ 统一使用像素值（数字类型），避免混用字符串和百分比
- ✅ 所有面板都设置合理的 min/max 约束
- ✅ 中间面板设置最小宽度，确保画布可用空间

### v4.3 (2025-10-27)

**滚动条和主题适配优化**（已废弃，v4.4版本改用更简单的方案）：
1. ~~**画布滚动条修复**：使用 Scroller 插件替代原生 panning，彻底移除浏览器滚动条，实现纯净画布~~（Scroller插件导致框选冲突）
2. **滚动条主题适配**：属性面板和逻辑积木面板滚动条完美适配暗色主题（8px细滚动条 + CSS变量）（✅ 保留）
3. **静态方法修复**：VisualEditorTab 使用 `App.useApp()` 替代静态 `message`，消除主题警告（✅ 保留）
4. ~~**操作优化**：画布平移改为直接拖拽空白区域（移除 Shift 键要求），操作更自然~~（v4.4改为Shift+拖拽平移，避免与框选冲突）

### v4.2 (2025-10-27)

**编辑器稳定性修复**：
1. 修复"点击节点消失"和"画布闪烁"bug（React hooks依赖链稳定性）
2. 使用 `useRef` 和 `useCallback` 稳定组件引用
3. 修复 Selection 插件配置错误

### v4.1 (2025-10-27)

**Bug修复与优化**：
1. 修复拖拽功能：使用原生HTML5 Drop API替代X6事件，添加onDragOver阻止默认行为
2. 后端新增NodeConfig.X/Y字段支持节点坐标持久化（types.go + lynxmanager.api）
3. 撤销功能优化：渲染时临时禁用历史记录，避免初始加载被记录
4. 状态管理优化：使用useMemo稳定actions引用、useRef追踪graphId避免重复初始化
5. 修复rc-collapse警告、manhattan路由错误（改用orth）
6. 文案统一：全局"逻辑块"改为"逻辑积木"

### v4.0 (2025-10-27)

**重大更新 - 可视化编辑器完整实现**：
1. 重构Tab2为完整三栏布局（左侧逻辑块面板/中间画布/右侧属性面板）
2. 实现拖拽添加节点、可视化连线、实时属性编辑
3. 集成X6插件（Snapline/History/Keyboard/Selection/Clipboard）
4. 实现自动布局算法（dagre）、撤销/重做、快捷键支持
5. 创建useGraphEditor Hook统一状态管理

**文件结构变更**：
- 新增 8 个组件文件（Toolbar、BlockPanel、GraphCanvas、PropertyPanel及样式）
- 新增 useGraphEditor Hook

**数据模型变更**：
- NodeConfig新增 `x` 和 `y` 字段（存储画布坐标）

**技术栈变更**：
- 新增依赖：dagre、@types/dagre、@antv/x6-plugin-*（6个插件）、nanoid

### v3.0 (2025-10-26)

**新增功能**：图标系统集成、图标选择器、卡片布局优化

**数据模型变更**：GraphConfigMetadata新增 `icon` 和 `iconColor` 字段

**技术栈变更**：新增 react-icons ^5.x

### v2.0 (2025-01-26)

**列表页架构调整**：从纯表格改为卡片+表格双视图，新增运行统计字段

### v1.0 (2025-10-26)

**初始版本**：基础表格列表 + 独立详情页架构

---

**文档结束**
