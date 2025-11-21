# Documentation 管理说明

> **说明**：本文档用于 Claude Code 理解如何管理项目文档和飞书文档的映射关系。

## 📂 核心文件

- `feishu-docs-mapping.json` - 飞书文档映射配置（关键文件）
- `Nexlyn 融合管理平台建设全家桶.md` - 项目管理主文档

## 🔍 查找飞书文档的方法

1. **读取本地Markdown顶部链接**（文件头部包含 `feishu_url` 和 `feishu_token`）
2. **读取 `feishu-docs-mapping.json`** 获取文档Token和URL
3. **格式**：`https://skylarkly.feishu.cn/docx/{TOKEN}`

## 📝 文档操作流程

### 创建新文档
```
1. 创建本地Markdown → 2. 导入飞书 → 3. 记录到mapping.json → 4. 在本地文档顶部添加链接
```

### 修改文档
- **推荐**：提醒用户直接在飞书中编辑
- **备选**：重新导入（会生成新文档，需更新链接）

## 🗂️ 全家桶文档子文档管理

### 概念
**全家桶文档** = 主索引文档 + 通过链接关联的所有子文档

### 子文档管理流程（4步）
```
1. 创建子文档（飞书或本地）
2. 更新 mapping.json（添加 parent_doc 字段）
3. 在全家桶主文档中添加链接
4. （可选）同步更新本地全家桶文档
```

### mapping.json 子文档格式
```json
{
  "子文档名称": {
    "local_path": "路径",
    "feishu_url": "https://skylarkly.feishu.cn/docx/xxxxx",
    "feishu_token": "xxxxx",
    "type": "docx",
    "description": "描述",
    "parent_doc": "全家桶文档",
    "created_at": "日期"
  }
}
```

### 子文档分类
| 分类 | 类型 | 飞书格式 |
|------|------|---------|
| 项目进度 | 甘特图、迭代计划 | 文档/表格/多维表格 |
| 重点事项 | 跟踪表、风险清单 | 多维表格 |
| 会议纪要 | 会议记录 | 文档 |
| 需求文档 | 功能需求、原型 | 文档/多维表格 |
| 问题清单 | Bug跟踪、技术债务 | 多维表格 |
| 项目资料 | 开发文档、培训 | 文档/知识库 |

### 命名规范
`[分类]-[模块]-[具体内容]`
例：`需求-物联管理-设备标签管理`

## 📌 当前文档

- **全家桶主文档**：https://skylarkly.feishu.cn/docx/LvyldTlbSoDGKCxSk5gcs9BXnFF

---

*最后更新：2025-01-04*
