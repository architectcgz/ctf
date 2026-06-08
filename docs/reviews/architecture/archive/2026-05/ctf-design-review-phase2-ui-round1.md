# CTF 平台前端设计 Review（phase2-ui 第 1 轮）：Phase 2 页面设计一致性审查

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | phase2-ui |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | design-system/pages/ 下 11 个 Phase 2 新增页面设计文件 |
| 变更概述 | T19~T30 任务对应的前端页面 UI 设计 |
| 审查基准 | design-system/MASTER.md + Phase 1 已有页面设计 |
| 审查日期 | 2026-03-02 |

---

## 审查结论

整体质量较好，设计风格与 MASTER.md 保持一致，组件复用合理。发现 7 个问题（2 高 / 3 中 / 2 低）。

---

## 问题清单

### 高-1：hint-system.md 使用 emoji 图标违反 MASTER 反模式规则

| 字段 | 说明 |
|------|------|
| 级别 | 高 |
| 文件 | hint-system.md |
| 位置 | wireframe 中 `✓ Hint 1`、`🔒 Hint 2`、`🔒 Hint 3` |

**问题描述：** MASTER.md 明确规定"禁止使用 emoji 作为功能图标，统一使用 Lucide Icons"。wireframe 中使用了 ✓ 和 🔒 emoji，虽然样式描述部分正确引用了 `CheckCircle` 和 `Lock` Lucide 图标，但 wireframe 本身会误导前端实现。

**修正建议：** wireframe 中统一使用 `[CheckCircle]` 和 `[Lock]` 标注替代 emoji，与样式描述保持一致。

---

### 高-2：progress-tracking.md 使用 emoji 火焰图标

| 字段 | 说明 |
|------|------|
| 级别 | 高 |
| 文件 | progress-tracking.md |
| 位置 | wireframe 中 `7 天 🔥` |

**问题描述：** 同上，违反 MASTER 反模式规则。样式描述中正确写了 `Lucide: Flame 20px text-warning`，但 wireframe 使用了 🔥 emoji。

**修正建议：** wireframe 改为 `7 天 [Flame]`。

---

### 中-1：topology-editor.md 拓扑编排器嵌入 Drawer 的空间问题

| 字段 | 说明 |
|------|------|
| 级别 | 中 |
| 文件 | topology-editor.md |
| 位置 | 第 1 节"入口" |

**问题描述：** 设计将拓扑编排器嵌入靶场创建/编辑 Drawer（admin.md 定义为 `w-[560px]`）。画布区域 + 右侧属性面板（`w-72` = 288px）在 560px 宽度内几乎无法使用，画布可用宽度仅约 272px。

**修正建议：** 两种方案：
1. 拓扑编排器使用全屏 overlay（`fixed inset-0 z-50 bg-base`）而非嵌入 Drawer
2. 将 Drawer 宽度扩展为 `w-full max-w-5xl`（1024px），仅拓扑步骤生效

推荐方案 1，全屏模式更适合画布类交互。

---

### 中-2：awd-contest.md 攻击面板 Flag 提交缺少频率限制说明

| 字段 | 说明 |
|------|------|
| 级别 | 中 |
| 文件 | awd-contest.md |
| 位置 | 第 4 节"攻击面板 > Flag 提交" |

**问题描述：** Phase 1 的 challenges.md 详情页明确定义了 Flag 提交频率限制的 UI 反馈（按钮禁用 + 剩余秒数），但 AWD 攻击面板的 Flag 提交未定义频率限制状态。AWD 模式下攻击频率更高，更需要此设计。

**修正建议：** 补充频率限制状态：
- 触发限制时：提交按钮禁用 + `text-warning text-xs` 显示 "冷却中 (Xs)"
- 与 challenges.md 保持一致的交互模式

---

### 中-3：靶场详情页信息架构膨胀

| 字段 | 说明 |
|------|------|
| 级别 | 中 |
| 文件 | hint-system.md、solve-timeline.md、writeup.md |
| 位置 | 三个文件均向 challenges.md 靶场详情页追加内容 |

**问题描述：** Phase 2 向靶场详情页叠加了三个新区块：增强提示系统（hint-system.md）、解题记录 Tab（solve-timeline.md）、Writeup 区域（writeup.md）。加上 Phase 1 已有的题目信息、实例控制、Flag 提交，单页信息密度过高。三个文件各自描述了追加方式，但缺少一个统一的靶场详情页 Phase 2 整合视图。

**修正建议：** 新增一个 `challenges-detail-v2.md` 整合文档，明确 Phase 2 后靶场详情页的完整布局：
- Tab 方案：`[题目详情] [解题记录] [Writeup]`
- 题目详情 Tab 内：题目信息 → 实例控制 + Flag 提交（并排）→ 提示系统
- 解题记录 Tab：时间线
- Writeup Tab：仅已解出/已公开时可见

---

### 低-1：awd-monitor.md 得分趋势 Tab 中"我方"措辞不适用于管理员视角

| 字段 | 说明 |
|------|------|
| 级别 | 低 |
| 文件 | awd-monitor.md |
| 位置 | 第 3 节"得分趋势 > 攻击力/防守力排行" |

**问题描述：** 排行列表中出现"我方"，但此页面是管理员监控面板，管理员不属于任何参赛队伍。"我方"是学员端 awd-contest.md 的概念，不应出现在管理员视角。

**修正建议：** 将示例数据中的"我方"替换为具体队伍名（如 "TeamF"）。

---

### 低-2：env-templates.md 缺少空状态设计

| 字段 | 说明 |
|------|------|
| 级别 | 低 |
| 文件 | env-templates.md |

**问题描述：** 模板列表页未定义空状态（无模板时的展示）。其他页面如 contest-export.md、solve-timeline.md 都有空状态设计，此处遗漏。

**修正建议：** 补充空状态：
```
居中:
  [LayoutTemplate 48px text-muted]
  "暂无环境模板"
  "在拓扑编排器中保存拓扑为模板，或导入已有模板"
  [+ 导入模板] 主按钮
```

---

## 跨页面一致性检查

### 通过项

| 检查项 | 结果 |
|--------|------|
| 色彩系统引用 | 全部正确引用 MASTER 色板，无自创色值 |
| 字体使用 | 数字/Flag/代码均使用 font-mono，正文使用默认无衬线 |
| 卡片样式 | 统一使用 bg-surface + border-border + rounded-lg |
| 按钮规范 | 主/次/危险按钮样式一致，尺寸标注清晰 |
| 表格规范 | 复用 MASTER 表格样式，表头/行/hover 一致 |
| 空状态 | 大部分页面有空状态设计（env-templates 除外，已列入问题） |
| 动画规范 | 遵循 150ms 微交互 / 300ms 排行榜变动，无过度动画 |
| 图标库 | 样式描述中统一使用 Lucide Icons（wireframe 有 emoji 问题已列入） |
| Dialog/Drawer 宽度 | 各处标注合理，无冲突 |
| 响应式 | 网格布局使用 grid-cols-1 md:2 lg:3，与 Phase 1 一致 |

### 设计亮点

1. **AWD 服务状态矩阵**（awd-monitor.md）— 轮次×队伍的热力矩阵直观展示全局态势，当前轮高亮列设计巧妙
2. **解题热力图**（progress-tracking.md）— GitHub 贡献图风格契合开发者用户群体，色阶使用主色 Teal 保持品牌一致
3. **网络层可视化**（topology-editor.md）— 三层网络用 cyan/amber/violet 半透明区域区分，既不抢眼又层次分明
4. **Writeup 剧透警告**（writeup.md）— 未解出但已公开时的折叠+警告设计，平衡了学习资源开放与解题体验
5. **提示扣分预览**（hint-system.md）— 解锁确认弹窗中的得分明细计算，降低用户决策成本

---

## 问题汇总

| # | 级别 | 文件 | 问题摘要 |
|---|------|------|----------|
| 高-1 | 高 | hint-system.md | wireframe 使用 emoji 图标 |
| 高-2 | 高 | progress-tracking.md | wireframe 使用 emoji 火焰 |
| 中-1 | 中 | topology-editor.md | 编排器嵌入 560px Drawer 空间不足 |
| 中-2 | 中 | awd-contest.md | 攻击面板缺少 Flag 提交频率限制 |
| 中-3 | 中 | hint/timeline/writeup | 靶场详情页缺少整合视图 |
| 低-1 | 低 | awd-monitor.md | 管理员视角出现"我方"措辞 |
| 低-2 | 低 | env-templates.md | 缺少空状态设计 |
